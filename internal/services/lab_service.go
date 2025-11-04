package services

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/s-samadi/ghas-lab-builder/internal/config"
	api "github.com/s-samadi/ghas-lab-builder/internal/github"
	userspec "github.com/s-samadi/ghas-lab-builder/internal/util"
)

func ProvisionOrgResources(workerId int, ctx context.Context, logger *slog.Logger, orgChan chan string, resultsChan chan string, enterprise *api.Enterprise, templateRepos []string, repoCounter *atomic.Int64, tokenManager *TokenManager) {

	logger.Info("Worker started", slog.Int("workerId", workerId))

	// Create a new organization for the user
	for user := range orgChan {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			logger.Warn("Worker stopping due to context cancellation", slog.Int("workerId", workerId))
			return
		default:
		}

		// Call the GraphQL-based CreateOrg function
		organization, err := enterprise.CreateOrg(ctx, logger, user)
		if err != nil {
			logger.Error("Failed to create organization",
				slog.String("user", user),
				slog.Any("error", err))
			continue
		}
		orgName := organization.Login

		logger.Info("Creating repositories in organization", slog.String("org", orgName))

		for _, repo := range templateRepos {
			logger.Info("Creating repository", slog.String("repo", repo))

			_, err := organization.CreateRepoFromTemplate(ctx, logger, repo)
			if err != nil {
				logger.Error("Failed to create repository",
					slog.String("repo", repo),
					slog.Any("error", err))
				continue
			}

			// Increment the repository counter
			count := repoCounter.Add(1)

			// Check if we need to rotate token (every 150 repos)
			if count%150 == 0 {
				logger.Info("Repository creation count reached 150, rotating token",
					slog.Int64("count", count),
					slog.Int("workerId", workerId))

				if err := tokenManager.RotateToken(); err != nil {
					logger.Error("Rate limit exceeded on all tokens. Please wait an hour before retrying.",
						slog.Int("tokens_used", tokenManager.GetTokenCount()),
						slog.Int("current_index", tokenManager.GetCurrentIndex()),
						slog.Any("error", err))
					return
				}
			}
		}

		resultsChan <- orgName
		logger.Info("Finished creating organization", slog.String("org", orgName))
	}

	logger.Info("Worker stopped", slog.Int("workerId", workerId))
}

func CreateLabEnvironment(ctx context.Context, logger *slog.Logger, usersFile string, templateReposFile string) error {

	//Get users
	logger.Info("Loading users from file", slog.String("file", usersFile))
	users, err := userspec.LoadFromFile(usersFile)
	if err != nil {
		return err
	}

	logger.Info("Loaded users", slog.Int("count", len(users)))

	templateRepos, err := userspec.LoadFromJsonFile(templateReposFile)
	if err != nil {
		return err
	}

	// Get enterprise slug from context
	enterpriseSlug, ok := ctx.Value(config.EnterpriseSlugKey).(string)
	if !ok {
		logger.Error("Enterprise slug not found in context")
		return err
	}

	// Create TokenManager
	tokenManager, err := NewTokenManager(ctx, logger)
	if err != nil {
		logger.Error("Failed to create token manager", slog.Any("error", err))
		return err
	}

	// Add TokenManager to context so it's available in CreateRepoFromTemplate
	ctx = context.WithValue(ctx, config.TokenManagerKey, tokenManager)

	//Get Enterprise details
	enterprise, err := api.GetEnterprise(ctx, logger, enterpriseSlug)
	if err != nil {
		logger.Error("Failed to get enterprise details", slog.String("slug", enterpriseSlug), slog.Any("error", err))
		return err
	}

	orgChan := make(chan string, len(users))
	resultsChan := make(chan string, len(users))

	// Create a thread-safe counter for repositories
	var repoCounter atomic.Int64

	// Use WaitGroup to track worker goroutines
	var wg sync.WaitGroup

	// Calculate optimal number of workers: min(9, number of users)
	numWorkers := 9
	if len(users) < numWorkers {
		numWorkers = len(users)
	}
	logger.Info("Starting workers", slog.Int("worker_count", numWorkers), slog.Int("user_count", len(users)))

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			ProvisionOrgResources(workerId, ctx, logger, orgChan, resultsChan, enterprise, templateRepos, &repoCounter, tokenManager)
		}(i)
	}

	// Send all users to the channel
	for _, user := range users {
		orgChan <- user
	}
	// Close orgChan immediately after sending all work
	close(orgChan)

	// Close resultsChan once all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	resultCount := 0

	for {
		select {
		case res, ok := <-resultsChan:
			if !ok {
				// Channel closed, all workers finished
				totalRepos := repoCounter.Load()
				if resultCount == len(users) {
					logger.Info("All organizations and repositories created successfully",
						slog.Int64("total_repos_created", totalRepos))
					return nil
				}
				logger.Error("Workers finished but not all users processed",
					slog.Int("expected", len(users)),
					slog.Int("processed", resultCount),
					slog.Int64("total_repos_created", totalRepos))
				return ctx.Err()
			}
			logger.Info("Created organization", slog.String("org", res))
			resultCount++
		case <-ctx.Done():
			logger.Error("Timeout reached while creating lab environment",
				slog.Int64("total_repos_created", repoCounter.Load()))
			return ctx.Err()
		}
	}
}

func DestroyOrgResources(workerId int, ctx context.Context, logger *slog.Logger, userChan chan string, resultsChan chan string, enterprise *api.Enterprise, labDate string) {
	logger.Info("Destroy worker started", slog.Int("workerId", workerId))

	for user := range userChan {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			logger.Warn("Destroy worker stopping due to context cancellation", slog.Int("workerId", workerId))
			return
		default:
		}

		orgName := "ghas-labs-" + labDate + "-" + user
		logger.Info("Deleting organization", slog.String("org", orgName), slog.String("user", user))

		// Call the GraphQL-based DeleteOrg function
		if err := enterprise.DeleteOrg(ctx, logger, orgName); err != nil {
			logger.Error("Failed to delete organization",
				slog.String("user", user),
				slog.String("org", orgName),
				slog.Any("error", err))
			// Still send result to avoid blocking
			resultsChan <- "failed:" + orgName
			continue
		}

		resultsChan <- orgName
		logger.Info("Finished deleting organization", slog.String("org", orgName))
	}

	logger.Info("Destroy worker stopped", slog.Int("workerId", workerId))
}

func DestroyLabEnvironment(ctx context.Context, logger *slog.Logger, labDate string, usersFile string) error {

	// Get users
	logger.Info("Loading users from file", slog.String("file", usersFile))
	users, err := userspec.LoadFromFile(usersFile)
	if err != nil {
		return err
	}

	logger.Info("Loaded users", slog.Int("count", len(users)))

	// Get enterprise slug from context
	enterpriseSlug, ok := ctx.Value(config.EnterpriseSlugKey).(string)
	if !ok {
		logger.Error("Enterprise slug not found in context")
		return err
	}

	// Get Enterprise details
	enterprise, err := api.GetEnterprise(ctx, logger, enterpriseSlug)
	if err != nil {
		logger.Error("Failed to get enterprise details", slog.String("slug", enterpriseSlug), slog.Any("error", err))
		return err
	}

	userChan := make(chan string, len(users))
	resultsChan := make(chan string, len(users))

	// Use WaitGroup to track worker goroutines
	var wg sync.WaitGroup

	// Calculate optimal number of workers: min(9, number of users)
	numWorkers := 9
	if len(users) < numWorkers {
		numWorkers = len(users)
	}
	logger.Info("Starting destroy workers", slog.Int("worker_count", numWorkers), slog.Int("user_count", len(users)))

	// Create worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			DestroyOrgResources(workerId, ctx, logger, userChan, resultsChan, enterprise, labDate)
		}(i)
	}

	// Send all users to the channel
	for _, user := range users {
		userChan <- user
	}
	// Close userChan immediately after sending all work
	close(userChan)

	// Close resultsChan once all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	resultCount := 0
	failedCount := 0

	for {
		select {
		case _, ok := <-resultsChan:
			if !ok {
				// Channel closed, all workers finished
				logger.Info("Finished destroying lab environment",
					slog.String("lab_date", labDate),
					slog.Int("total", len(users)),
					slog.Int("processed", resultCount),
					slog.Int("failed", failedCount))

				if failedCount > 0 {
					return ctx.Err()
				}
				return nil
			}

			resultCount++
		case <-ctx.Done():
			logger.Error("Timeout reached while destroying lab environment")
			return ctx.Err()
		}
	}
}
