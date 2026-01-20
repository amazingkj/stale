package email

import (
	"strings"
	"testing"

	"github.com/jiin/stale/internal/domain"
)

func TestNew(t *testing.T) {
	service := New()

	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestSendNewOutdatedReport_EmailDisabled(t *testing.T) {
	service := New()
	settings := &domain.Settings{
		EmailEnabled: false,
	}
	report := &domain.NewOutdatedReport{
		NewOutdated: []domain.DependencyWithRepo{
			{Dependency: domain.Dependency{Name: "test-dep"}, RepoFullName: "owner/repo"},
		},
	}

	err := service.SendNewOutdatedReport(settings, report)
	if err != nil {
		t.Errorf("expected no error when email disabled, got %v", err)
	}
}

func TestSendNewOutdatedReport_EmptyReport(t *testing.T) {
	service := New()
	settings := &domain.Settings{
		EmailEnabled: true,
	}
	report := &domain.NewOutdatedReport{
		NewOutdated: []domain.DependencyWithRepo{},
	}

	err := service.SendNewOutdatedReport(settings, report)
	if err != nil {
		t.Errorf("expected no error for empty report, got %v", err)
	}
}

func TestBuildEmailBody(t *testing.T) {
	service := New()
	report := &domain.NewOutdatedReport{
		ScanID: 123,
		NewOutdated: []domain.DependencyWithRepo{
			{
				Dependency:   domain.Dependency{Name: "react", CurrentVersion: "17.0.0", LatestVersion: "18.2.0", Ecosystem: "npm"},
				RepoFullName: "owner/frontend",
			},
			{
				Dependency:   domain.Dependency{Name: "spring-boot", CurrentVersion: "2.7.0", LatestVersion: "3.1.0", Ecosystem: "maven"},
				RepoFullName: "owner/backend",
			},
		},
	}

	body, err := service.buildEmailBody(report)
	if err != nil {
		t.Fatalf("buildEmailBody failed: %v", err)
	}

	// Check that the body contains expected elements
	expectedStrings := []string{
		"New Outdated Dependencies Found",
		"2 new outdated dependencies",
		"scan #123",
		"react",
		"owner/frontend",
		"17.0.0",
		"18.2.0",
		"npm",
		"spring-boot",
		"owner/backend",
		"2.7.0",
		"3.1.0",
		"maven",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(body, s) {
			t.Errorf("expected body to contain %q", s)
		}
	}
}

func TestBuildEmailBody_SingleDependency(t *testing.T) {
	service := New()
	report := &domain.NewOutdatedReport{
		ScanID: 1,
		NewOutdated: []domain.DependencyWithRepo{
			{
				Dependency:   domain.Dependency{Name: "lodash", CurrentVersion: "4.0.0", LatestVersion: "4.17.21", Ecosystem: "npm"},
				RepoFullName: "owner/app",
			},
		},
	}

	body, err := service.buildEmailBody(report)
	if err != nil {
		t.Fatalf("buildEmailBody failed: %v", err)
	}

	if !strings.Contains(body, "1 new outdated dependencies") {
		t.Error("expected body to contain '1 new outdated dependencies'")
	}
	if !strings.Contains(body, "lodash") {
		t.Error("expected body to contain 'lodash'")
	}
}

func TestBuildEmailBody_HTMLStructure(t *testing.T) {
	service := New()
	report := &domain.NewOutdatedReport{
		ScanID: 1,
		NewOutdated: []domain.DependencyWithRepo{
			{Dependency: domain.Dependency{Name: "test", CurrentVersion: "1.0.0", LatestVersion: "2.0.0", Ecosystem: "npm"}, RepoFullName: "owner/repo"},
		},
	}

	body, err := service.buildEmailBody(report)
	if err != nil {
		t.Fatalf("buildEmailBody failed: %v", err)
	}

	// Check HTML structure
	requiredTags := []string{
		"<!DOCTYPE html>",
		"<html>",
		"</html>",
		"<head>",
		"</head>",
		"<body>",
		"</body>",
		"<style>",
		"</style>",
		"<table>",
		"</table>",
	}

	for _, tag := range requiredTags {
		if !strings.Contains(body, tag) {
			t.Errorf("expected body to contain %q", tag)
		}
	}
}

func TestBuildEmailBody_EcosystemStyles(t *testing.T) {
	service := New()

	ecosystems := []string{"npm", "maven", "gradle", "go"}

	for _, ecosystem := range ecosystems {
		report := &domain.NewOutdatedReport{
			ScanID: 1,
			NewOutdated: []domain.DependencyWithRepo{
				{Dependency: domain.Dependency{Name: "test", CurrentVersion: "1.0.0", LatestVersion: "2.0.0", Ecosystem: ecosystem}, RepoFullName: "owner/repo"},
			},
		}

		body, err := service.buildEmailBody(report)
		if err != nil {
			t.Fatalf("buildEmailBody failed for ecosystem %s: %v", ecosystem, err)
		}

		// Check that ecosystem class is present in both style and content
		if !strings.Contains(body, "."+ecosystem) {
			t.Errorf("expected style for ecosystem %s", ecosystem)
		}
	}
}

func TestRecipientParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"user@example.com", []string{"user@example.com"}},
		{"user1@example.com,user2@example.com", []string{"user1@example.com", "user2@example.com"}},
		{"user1@example.com, user2@example.com", []string{"user1@example.com", "user2@example.com"}},
		{"  user@example.com  ", []string{"user@example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			recipients := strings.Split(tt.input, ",")
			for i, r := range recipients {
				recipients[i] = strings.TrimSpace(r)
			}

			if len(recipients) != len(tt.expected) {
				t.Errorf("expected %d recipients, got %d", len(tt.expected), len(recipients))
			}

			for i, r := range recipients {
				if r != tt.expected[i] {
					t.Errorf("recipient %d: expected %q, got %q", i, tt.expected[i], r)
				}
			}
		})
	}
}
