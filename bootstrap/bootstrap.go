package bootstrap

import (
	"context"
	"fmt"
	"kaizen-hq/internal/permission"
	"kaizen-hq/internal/role"
	"kaizen-hq/internal/account"
	"log"
	"os"

	"golang.org/x/term"
)

// * Seedsystem populates the database when first created with admin data
func SeedSystem(
	ctx context.Context,
	userSvc *user.Service,
	roleSvc *role.Service,
	permSvc *permission.Service,
) error {
	// check if any users exist
	userCount, err := userSvc.Count(ctx)
	if err != nil {
		return err
	}
	if userCount > 0 {
		log.Println("Bootstrap skipped: users already exist")
		return nil
	}
	log.Println("Bootstrapping system...")

	var adminEmail string
	var adminTornID int
	var adminAPIKey string

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

	fmt.Print("Enter admin email: ")
	fmt.Scanln(&adminEmail)

	fmt.Print("Enter password: ")
	passwordBytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
	adminPassword := string(passwordBytes)

	fmt.Print("Enter admin's torn ID: ")
	fmt.Scanln(&adminTornID)

	fmt.Print("Enter admin's api key: ")
	fmt.Scanln(&adminAPIKey)

	// Create root user
	userID, err := userSvc.CreateUser(ctx, &user.User{
		Email:    adminEmail,
		TornID:   adminTornID,
		Password: adminPassword,
		APIKey:   adminAPIKey,
	})

	if err != nil {
		return err
	}

	// Assign admin role
	_ = userSvc.AssignRole(ctx, userID, adminRole.ID)

	return nil
}
