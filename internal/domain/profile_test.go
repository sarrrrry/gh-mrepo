package domain_test

import (
	"testing"

	"github.com/sarrrrry/gh-mrepo/internal/domain"
)

func TestNewProfile_Success(t *testing.T) {
	p, err := domain.NewProfile("work", "/home/user/.config/gh", "/home/user/repos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "work" {
		t.Errorf("Name = %q, want %q", p.Name, "work")
	}
	if p.GHConfigDir != "/home/user/.config/gh" {
		t.Errorf("GHConfigDir = %q, want %q", p.GHConfigDir, "/home/user/.config/gh")
	}
	if p.Root != "/home/user/repos" {
		t.Errorf("Root = %q, want %q", p.Root, "/home/user/repos")
	}
}

func TestNewProfile_EmptyRoot(t *testing.T) {
	p, err := domain.NewProfile("personal", "/home/user/.config/gh", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Root != "" {
		t.Errorf("Root = %q, want empty", p.Root)
	}
}

func TestNewProfile_EmptyGHConfigDir(t *testing.T) {
	_, err := domain.NewProfile("work", "", "/home/user/repos")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != domain.ErrEmptyGHConfigDir {
		t.Errorf("err = %v, want %v", err, domain.ErrEmptyGHConfigDir)
	}
}

func TestNewProfile_EmptyName(t *testing.T) {
	_, err := domain.NewProfile("", "/home/user/.config/gh", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != domain.ErrEmptyName {
		t.Errorf("err = %v, want %v", err, domain.ErrEmptyName)
	}
}
