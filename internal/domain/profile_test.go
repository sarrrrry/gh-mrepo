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

func TestProfile_GitConfigFields(t *testing.T) {
	p, err := domain.NewProfile("work", "/home/user/.config/gh", "/home/user/repos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p.GitConfigName = "Work User"
	p.GitConfigEmail = "work@example.com"

	if p.GitConfigName != "Work User" {
		t.Errorf("GitConfigName = %q, want %q", p.GitConfigName, "Work User")
	}
	if p.GitConfigEmail != "work@example.com" {
		t.Errorf("GitConfigEmail = %q, want %q", p.GitConfigEmail, "work@example.com")
	}
}

func TestProfile_GitConfigFieldsEmpty(t *testing.T) {
	p, err := domain.NewProfile("personal", "/home/user/.config/gh", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.GitConfigName != "" {
		t.Errorf("GitConfigName = %q, want empty", p.GitConfigName)
	}
	if p.GitConfigEmail != "" {
		t.Errorf("GitConfigEmail = %q, want empty", p.GitConfigEmail)
	}
}

func TestFindByDirectory_Match(t *testing.T) {
	profiles := []domain.Profile{
		{Name: "personal", GHConfigDir: "/config/personal", Root: "/home/user/personal"},
		{Name: "work", GHConfigDir: "/config/work", Root: "/home/user/work"},
	}

	p, err := domain.FindByDirectory(profiles, "/home/user/work/some-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "work" {
		t.Errorf("Name = %q, want %q", p.Name, "work")
	}
}

func TestFindByDirectory_FirstMatch(t *testing.T) {
	profiles := []domain.Profile{
		{Name: "parent", GHConfigDir: "/config/parent", Root: "/home/user"},
		{Name: "child", GHConfigDir: "/config/child", Root: "/home/user/child"},
	}

	p, err := domain.FindByDirectory(profiles, "/home/user/child/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "parent" {
		t.Errorf("Name = %q, want %q (first match wins)", p.Name, "parent")
	}
}

func TestFindByDirectory_NoMatch(t *testing.T) {
	profiles := []domain.Profile{
		{Name: "work", GHConfigDir: "/config/work", Root: "/home/user/work"},
	}

	_, err := domain.FindByDirectory(profiles, "/tmp/other")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFindByDirectory_SkipsEmptyRoot(t *testing.T) {
	profiles := []domain.Profile{
		{Name: "noroot", GHConfigDir: "/config/noroot", Root: ""},
		{Name: "work", GHConfigDir: "/config/work", Root: "/home/user/work"},
	}

	p, err := domain.FindByDirectory(profiles, "/home/user/work/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "work" {
		t.Errorf("Name = %q, want %q", p.Name, "work")
	}
}
