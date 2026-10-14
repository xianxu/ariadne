package main

import "context"

// Context-free aliases are fixture helpers; production always carries the command context.
func archiveRead(root string, args ...string) ([]byte, error) {
	return archiveReadContext(context.Background(), root, args...)
}

func archiveTree(root, ref string, dirs landingArchiveRoots) (landingArchiveSnapshot, error) {
	return archiveTreeContext(context.Background(), root, ref, dirs)
}

func archiveBlob(root string, b landingArchiveBlob) ([]byte, error) {
	return archiveBlobContext(context.Background(), root, b)
}

func archiveSnapshot(root, ref string, selected []landingOwnedIssue, dirs landingArchiveRoots) (landingArchiveSnapshot, error) {
	return archiveSnapshotContext(context.Background(), root, ref, selected, dirs)
}
