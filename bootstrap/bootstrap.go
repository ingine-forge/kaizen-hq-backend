package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"kaizen-hq/internal/account"
	"kaizen-hq/internal/client"
	"kaizen-hq/internal/permission"
	"kaizen-hq/internal/role"
	"kaizen-hq/internal/user"
	"log"
	"os"

	"golang.org/x/term"
)

// * Seedsystem populates the database when first created with admin data
func SeedSystem(
	ctx context.Context,
	accountSvc *account.Service,
	userSvc *user.Service,
	roleSvc *role.Service,
	permSvc *permission.Service,
) error {
	// check if any users exist
	if accountSvc == nil {
		return fmt.Errorf("account service is nil")
	}
	accountCount, err := accountSvc.Count(ctx)
	if err != nil {
		return err
	}
	if accountCount > 0 {
		log.Println("Bootstrap skipped: users already exist")
		return nil
	}
	log.Println("Bootstrapping system...")

	var adminEmail string
	var adminAPIKey string

	fmt.Print("Enter admin email: ")
	fmt.Scanln(&adminEmail)

	fmt.Print("Enter password: ")
	passwordBytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
	adminPassword := string(passwordBytes)

	fmt.Println()

	fmt.Print("Enter admin's api key: ")
	fmt.Scanln(&adminAPIKey)

	tornClient := client.NewClient()

	user, err := tornClient.FetchTornUser(ctx, adminAPIKey, "")

	if err != nil {
		return err
	}

	discordID, err := tornClient.FetchDiscordID(ctx, adminAPIKey, user.PlayerID)

	if err != nil {
		discordID = ""
		return err
	}

	err = userSvc.CreateUser(ctx, user.PlayerID, adminAPIKey)
	if err != nil {
		return errors.New("error creating the user")
	}

	// Create root user
	accountID, err := accountSvc.CreateAccount(ctx, &account.Account{
		Email:     adminEmail,
		TornID:    user.PlayerID,
		Password:  adminPassword,
		APIKey:    adminAPIKey,
		DiscordID: discordID,
	})

	if err != nil {
		return err
	}

	// Step 1: Create roles
	adminRole, err := roleSvc.Create(ctx, &role.Role{Name: "admin", Description: "Full access"})
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Step 2: Create permissions
	logPerm, err := permSvc.Create(ctx, &permission.Permission{Name: "view_logs", Description: "Able to view logs"})
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Step 3: Assign permission to role
	err = roleSvc.AssignPermission(ctx, adminRole, logPerm)

	if err != nil {
		fmt.Println(err)
	}

	// Assign admin role
	_ = accountSvc.AssignRole(ctx, accountID, adminRole.ID)

	return nil
}
