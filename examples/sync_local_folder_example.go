package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andreisuslov/cloud-sync/pkg/backup"
)

func main() {
	// Example: Configure and sync local folders with cloud storage

	// Step 1: Create backup manager
	config := &backup.Config{
		Username:     os.Getenv("USER"),
		HomeDir:      os.Getenv("HOME"),
		SourceRemote: "backblaze",  // Your source remote name
		SourceBucket: "my-bucket",  // Your source bucket
		DestRemote:   "s3",         // Your destination remote name
		DestBucket:   "backup-bucket", // Your destination bucket
	}

	manager, err := backup.NewManager(config)
	if err != nil {
		log.Fatalf("Failed to create backup manager: %v", err)
	}

	// Step 2: Add sync pairs for local folders

	// Example 1: Upload documents to cloud
	fmt.Println("Adding Documents sync pair...")
	err = manager.AddSyncPair(
		"documents-backup",                    // Name
		os.Getenv("HOME")+"/Documents",        // Local path
		"backblaze",                           // Remote name
		"my-bucket/documents",                 // Remote path
		"upload",                              // Direction: local → remote
	)
	if err != nil {
		log.Printf("Warning: Could not add documents sync pair: %v", err)
	}

	// Example 2: Download photos from cloud
	fmt.Println("Adding Photos sync pair...")
	err = manager.AddSyncPair(
		"photos-download",
		os.Getenv("HOME")+"/Pictures/CloudPhotos",
		"s3",
		"photo-bucket/archive",
		"download", // Direction: remote → local
	)
	if err != nil {
		log.Printf("Warning: Could not add photos sync pair: %v", err)
	}

	// Example 3: Bidirectional sync for work folder
	fmt.Println("Adding Work sync pair...")
	err = manager.AddSyncPair(
		"work-sync",
		os.Getenv("HOME")+"/Work",
		"gdrive",
		"work-backup",
		"bidirectional", // Direction: both ways
	)
	if err != nil {
		log.Printf("Warning: Could not add work sync pair: %v", err)
	}

	// Step 3: List all configured sync pairs
	fmt.Println("\n=== Configured Sync Pairs ===")
	pairs, err := manager.ListSyncPairs()
	if err != nil {
		log.Fatalf("Failed to list sync pairs: %v", err)
	}

	for i, pair := range pairs {
		status := "✓ Enabled"
		if !pair.Enabled {
			status = "✗ Disabled"
		}
		fmt.Printf("%d. [%s] %s\n", i+1, status, pair.Name)
		fmt.Printf("   Local:  %s\n", pair.LocalPath)
		fmt.Printf("   Remote: %s:%s\n", pair.RemoteName, pair.RemotePath)
		fmt.Printf("   Direction: %s\n\n", pair.Direction)
	}

	// Step 4: Sync a specific folder (with dry-run)
	fmt.Println("=== Syncing Documents (Dry Run) ===")
	err = manager.SyncPair("documents-backup", true, true) // progress=true, dryRun=true
	if err != nil {
		log.Printf("Dry run failed: %v", err)
	}

	// Step 5: Sync all enabled folders (commented out for safety)
	// Uncomment to actually sync all folders
	/*
	fmt.Println("\n=== Syncing All Enabled Folders ===")
	err = manager.SyncAllEnabled(true, false) // progress=true, dryRun=false
	if err != nil {
		log.Fatalf("Sync failed: %v", err)
	}
	*/

	// Step 6: Toggle a sync pair
	fmt.Println("\n=== Toggling Documents Sync Pair ===")
	err = manager.ToggleSyncPair("documents-backup")
	if err != nil {
		log.Printf("Toggle failed: %v", err)
	}

	// Step 7: Remove a sync pair (commented out for safety)
	/*
	fmt.Println("\n=== Removing Work Sync Pair ===")
	err = manager.RemoveSyncPair("work-sync")
	if err != nil {
		log.Printf("Remove failed: %v", err)
	}
	*/

	fmt.Println("\n✓ Example completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Configure your rclone remotes: rclone config")
	fmt.Println("2. Adjust the sync pairs above to match your setup")
	fmt.Println("3. Run with dry-run first to preview changes")
	fmt.Println("4. Enable actual syncing by setting dryRun=false")
}
