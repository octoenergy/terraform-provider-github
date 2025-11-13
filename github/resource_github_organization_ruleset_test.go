package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-github/v67/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestGithubOrganizationRulesets(t *testing.T) {
	if isEnterprise != "true" {
		t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
	}

	if testEnterprise == "" {
		t.Skip("Skipping because `ENTERPRISE_SLUG` is not set")
	}

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("Creates and updates organization rulesets without errors", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					creation = true

					update = true

					deletion                = true
					required_linear_history = true

					required_signatures = false

					pull_request {
						required_approving_review_count   = 2
						required_review_thread_resolution = true
						require_code_owner_review         = true
						dismiss_stale_reviews_on_push     = true
						require_last_push_approval        = true
					}

					required_status_checks {

						required_check {
							context = "ci"
						}

						strict_required_status_checks_policy = true
						do_not_enforce_on_create             = true
					}

					required_workflows {
						do_not_enforce_on_create = true
						required_workflow {
							path          = "path/to/workflow.yaml"
							repository_id = 1234
						}
					}

					required_code_scanning {
					  required_code_scanning_tool {
						alerts_threshold = "errors"
						security_alerts_threshold = "high_or_higher"
						tool = "CodeQL"
					  }
					}

					branch_name_pattern {
						name     = "test"
						negate   = false
						operator = "starts_with"
						pattern  = "test"
					}

					non_fast_forward = true
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"name",
				"test",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"enforcement",
				"active",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"rules.0.required_workflows.0.do_not_enforce_on_create",
				"true",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"rules.0.required_workflows.0.required_workflow.0.path",
				"path/to/workflow.yaml",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"rules.0.required_workflows.0.required_workflow.0.repository_id",
				"1234",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"rules.0.required_code_scanning.0.required_code_scanning_tool.0.alerts_threshold",
				"errors",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"rules.0.required_code_scanning.0.required_code_scanning_tool.0.security_alerts_threshold",
				"high_or_higher",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"rules.0.required_code_scanning.0.required_code_scanning_tool.0.tool",
				"CodeQL",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Updates a ruleset name without error", func(t *testing.T) {

		oldRSName := fmt.Sprintf(`ruleset-%[1]s`, randomID)
		newRSName := fmt.Sprintf(`%[1]s-renamed`, randomID)

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "%s"
				target      = "branch"
				enforcement = "active"

				rules {
					creation = true
				}
			}
		`, oldRSName)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_organization_ruleset.test", "name",
					oldRSName,
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_organization_ruleset.test", "name",
					newRSName,
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						// Rename the ruleset to something else
						Config: strings.Replace(
							config,
							oldRSName,
							newRSName, 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Imports rulesets without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					creation = true

					update = true

					deletion                = true
					required_linear_history = true

					required_signatures = false

					pull_request {
						required_approving_review_count   = 2
						required_review_thread_resolution = true
						require_code_owner_review         = true
						dismiss_stale_reviews_on_push     = true
						require_last_push_approval        = true
					}

					required_status_checks {

						required_check {
							context = "ci"
						}

						strict_required_status_checks_policy = true
						do_not_enforce_on_create             = true
					}

					branch_name_pattern {
						name     = "test"
						negate   = false
						operator = "starts_with"
						pattern  = "test"
					}

					non_fast_forward = true
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_organization_ruleset.test", "name"),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
					{
						ResourceName:      "github_organization_ruleset.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Creates and updates organization using bypasses", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-%s"
				target      = "branch"
				enforcement = "active"

				bypass_actors {
					actor_type = "DeployKey"
					bypass_mode = "always"
				}

				bypass_actors {
					actor_id    = 5
					actor_type  = "RepositoryRole"
					bypass_mode = "always"
				}

				bypass_actors {
					actor_id    = 1
					actor_type  = "OrganizationAdmin"
					bypass_mode = "always"
				}

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					creation = true
					update = true
					deletion                = true
					required_linear_history = true
					required_signatures = false
					pull_request {
						required_approving_review_count   = 2
						required_review_thread_resolution = true
						require_code_owner_review         = true
						dismiss_stale_reviews_on_push     = true
						require_last_push_approval        = true
					}
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.#",
				"3",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.0.actor_type",
				"DeployKey",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.0.bypass_mode",
				"always",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.1.actor_id",
				"5",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.1.actor_type",
				"RepositoryRole",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.1.bypass_mode",
				"always",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.2.actor_id",
				"1",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.2.actor_type",
				"OrganizationAdmin",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.2.bypass_mode",
				"always",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Creates organization ruleset with all bypass_modes", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-bypass-modes-%s"
				target      = "branch"
				enforcement = "active"

				bypass_actors {
					actor_id    = 1
					actor_type  = "OrganizationAdmin"
					bypass_mode = "always"
				}

				bypass_actors {
					actor_id    = 5
					actor_type  = "RepositoryRole"
					bypass_mode = "pull_request"
				}

				bypass_actors {
					actor_id    = 2
					actor_type  = "RepositoryRole"
					bypass_mode = "exempt"
				}

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					creation = true
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.#",
				"3",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.0.actor_id",
				"1",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.0.actor_type",
				"OrganizationAdmin",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.0.bypass_mode",
				"always",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.1.actor_id",
				"5",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.1.actor_type",
				"RepositoryRole",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.1.bypass_mode",
				"pull_request",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.2.actor_id",
				"2",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.2.actor_type",
				"RepositoryRole",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test", "bypass_actors.2.bypass_mode",
				"exempt",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Updates organization ruleset bypass_mode without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-bypass-update-%s"
				target      = "branch"
				enforcement = "active"

				bypass_actors {
					actor_id    = 1
					actor_type  = "OrganizationAdmin"
					bypass_mode = "always"
				}

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					creation = true
				}
			}
		`, randomID)

		configUpdated := strings.Replace(
			config,
			`bypass_mode = "always"`,
			`bypass_mode = "exempt"`,
			1,
		)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_organization_ruleset.test", "bypass_actors.0.bypass_mode",
					"always",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_organization_ruleset.test", "bypass_actors.0.bypass_mode",
					"exempt",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: configUpdated,
						Check:  checks["after"],
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

}

func TestOrganizationRulesetWithRepositoryProperty(t *testing.T) {
	if isEnterprise != "true" {
		t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
	}

	if testEnterprise == "" {
		t.Skip("Skipping because `ENTERPRISE_SLUG` is not set")
	}

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("Creates organization ruleset with repository_property conditions", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-repo-props-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~DEFAULT_BRANCH"]
						exclude = []
					}

					repository_property {
						include {
							name            = "environment"
							property_values = ["production", "staging"]
						}

						include {
							name            = "compliance"
							property_values = ["high"]
							source          = "custom"
						}

						exclude {
							name            = "archived"
							property_values = ["true"]
						}
					}
				}

				rules {
					pull_request {
						required_approving_review_count = 2
					}
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"name",
				fmt.Sprintf("test-repo-props-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"enforcement",
				"active",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.include.0.name",
				"environment",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.include.0.property_values.0",
				"production",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.include.0.property_values.1",
				"staging",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.include.1.name",
				"compliance",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.include.1.property_values.0",
				"high",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.include.1.source",
				"custom",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.exclude.0.name",
				"archived",
			),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.exclude.0.property_values.0",
				"true",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Updates organization ruleset repository_property without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-repo-props-update-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~DEFAULT_BRANCH"]
						exclude = []
					}

					repository_property {
						include {
							name            = "environment"
							property_values = ["production"]
						}
					}
				}

				rules {
					creation = true
				}
			}
		`, randomID)

		configUpdated := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-repo-props-update-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~DEFAULT_BRANCH"]
						exclude = []
					}

					repository_property {
						include {
							name            = "environment"
							property_values = ["production", "staging"]
						}

						exclude {
							name            = "deprecated"
							property_values = ["true"]
						}
					}
				}

				rules {
					creation = true
				}
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_organization_ruleset.test",
					"conditions.0.repository_property.0.include.0.property_values.#",
					"1",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_organization_ruleset.test",
					"conditions.0.repository_property.0.include.0.property_values.#",
					"2",
				),
				resource.TestCheckResourceAttr(
					"github_organization_ruleset.test",
					"conditions.0.repository_property.0.exclude.0.name",
					"deprecated",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: configUpdated,
						Check:  checks["after"],
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Imports organization ruleset with repository_property without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-repo-props-import-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~DEFAULT_BRANCH"]
						exclude = []
					}

					repository_property {
						include {
							name            = "tier"
							property_values = ["premium"]
						}
					}
				}

				rules {
					creation = true
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_organization_ruleset.test", "name"),
			resource.TestCheckResourceAttr(
				"github_organization_ruleset.test",
				"conditions.0.repository_property.0.include.0.name",
				"tier",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
					{
						ResourceName:      "github_organization_ruleset.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})
}

func TestExpandFlattenRepositoryPropertyConditions(t *testing.T) {
	// Unit test for repository_property expand/flatten functionality

	t.Run("expands repository property conditions correctly", func(t *testing.T) {
		conditionsMap := map[string]interface{}{
			"ref_name": []interface{}{
				map[string]interface{}{
					"include": []interface{}{"~DEFAULT_BRANCH"},
					"exclude": []interface{}{},
				},
			},
			"repository_property": []interface{}{
				map[string]interface{}{
					"include": []interface{}{
						map[string]interface{}{
							"name":            "environment",
							"property_values": []interface{}{"production", "staging"},
						},
						map[string]interface{}{
							"name":            "compliance",
							"property_values": []interface{}{"high"},
							"source":          "custom",
						},
					},
					"exclude": []interface{}{
						map[string]interface{}{
							"name":            "archived",
							"property_values": []interface{}{"true"},
						},
					},
				},
			},
		}

		input := []interface{}{conditionsMap}

		// Test expand functionality (organization rulesets use org=true)
		expandedConditions := expandConditions(input, true)

		if expandedConditions == nil {
			t.Fatal("Expected expanded conditions, got nil")
		}

		if expandedConditions.RepositoryProperty == nil {
			t.Fatal("Expected repository_property to be expanded, got nil")
		}

		// Verify include conditions
		if len(expandedConditions.RepositoryProperty.Include) != 2 {
			t.Fatalf("Expected 2 include conditions, got %d", len(expandedConditions.RepositoryProperty.Include))
		}

		// Check first include condition
		firstInclude := expandedConditions.RepositoryProperty.Include[0]
		if firstInclude.Name != "environment" {
			t.Errorf("Expected first include name to be 'environment', got '%s'", firstInclude.Name)
		}
		if len(firstInclude.Values) != 2 {
			t.Errorf("Expected 2 property values in first include, got %d", len(firstInclude.Values))
		}
		if firstInclude.Values[0] != "production" || firstInclude.Values[1] != "staging" {
			t.Errorf("Expected property values ['production', 'staging'], got %v", firstInclude.Values)
		}

		// Check second include condition
		secondInclude := expandedConditions.RepositoryProperty.Include[1]
		if secondInclude.Name != "compliance" {
			t.Errorf("Expected second include name to be 'compliance', got '%s'", secondInclude.Name)
		}
		if secondInclude.Source == nil || *secondInclude.Source != "custom" {
			t.Errorf("Expected source to be 'custom', got %v", secondInclude.Source)
		}

		// Verify exclude conditions
		if len(expandedConditions.RepositoryProperty.Exclude) != 1 {
			t.Fatalf("Expected 1 exclude condition, got %d", len(expandedConditions.RepositoryProperty.Exclude))
		}

		firstExclude := expandedConditions.RepositoryProperty.Exclude[0]
		if firstExclude.Name != "archived" {
			t.Errorf("Expected exclude name to be 'archived', got '%s'", firstExclude.Name)
		}
	})

	t.Run("flattens repository property conditions correctly", func(t *testing.T) {
		// Create test data using the github library structs
		source := "custom"
		conditions := &github.RulesetConditions{
			RefName: &github.RulesetRefConditionParameters{
				Include: []string{"~DEFAULT_BRANCH"},
				Exclude: []string{},
			},
			RepositoryProperty: &github.RulesetRepositoryPropertyConditionParameters{
				Include: []github.RulesetRepositoryPropertyTargetParameters{
					{
						Name:   "environment",
						Values: []string{"production", "staging"},
					},
					{
						Name:   "compliance",
						Values: []string{"high"},
						Source: &source,
					},
				},
				Exclude: []github.RulesetRepositoryPropertyTargetParameters{
					{
						Name:   "archived",
						Values: []string{"true"},
					},
				},
			},
		}

		// Test flatten functionality (organization rulesets use org=true)
		flattenedResult := flattenConditions(conditions, true)

		if len(flattenedResult) != 1 {
			t.Fatalf("Expected 1 flattened result, got %d", len(flattenedResult))
		}

		flattenedConditionsMap := flattenedResult[0].(map[string]interface{})

		// Verify repository_property exists
		repositoryPropertySlice, ok := flattenedConditionsMap["repository_property"].([]map[string]interface{})
		if !ok {
			t.Fatal("Expected repository_property to be present in flattened conditions")
		}
		if len(repositoryPropertySlice) != 1 {
			t.Fatalf("Expected 1 repository_property block, got %d", len(repositoryPropertySlice))
		}

		repositoryProperty := repositoryPropertySlice[0]

		// Verify include conditions
		includeSlice, ok := repositoryProperty["include"].([]map[string]interface{})
		if !ok {
			t.Fatal("Expected include to be present")
		}
		if len(includeSlice) != 2 {
			t.Fatalf("Expected 2 include conditions, got %d", len(includeSlice))
		}

		// Check first include
		if includeSlice[0]["name"] != "environment" {
			t.Errorf("Expected first include name to be 'environment', got '%v'", includeSlice[0]["name"])
		}
		propertyValues := includeSlice[0]["property_values"].([]string)
		if len(propertyValues) != 2 {
			t.Errorf("Expected 2 property values, got %d", len(propertyValues))
		}

		// Check second include with source
		if includeSlice[1]["name"] != "compliance" {
			t.Errorf("Expected second include name to be 'compliance', got '%v'", includeSlice[1]["name"])
		}
		if includeSlice[1]["source"] != "custom" {
			t.Errorf("Expected source to be 'custom', got '%v'", includeSlice[1]["source"])
		}

		// Verify exclude conditions
		excludeSlice, ok := repositoryProperty["exclude"].([]map[string]interface{})
		if !ok {
			t.Fatal("Expected exclude to be present")
		}
		if len(excludeSlice) != 1 {
			t.Fatalf("Expected 1 exclude condition, got %d", len(excludeSlice))
		}
		if excludeSlice[0]["name"] != "archived" {
			t.Errorf("Expected exclude name to be 'archived', got '%v'", excludeSlice[0]["name"])
		}
	})

	t.Run("round-trip test for repository property conditions", func(t *testing.T) {
		// Test that expand -> flatten returns the same data
		originalMap := map[string]interface{}{
			"ref_name": []interface{}{
				map[string]interface{}{
					"include": []interface{}{"~DEFAULT_BRANCH"},
					"exclude": []interface{}{},
				},
			},
			"repository_property": []interface{}{
				map[string]interface{}{
					"include": []interface{}{
						map[string]interface{}{
							"name":            "tier",
							"property_values": []interface{}{"premium", "enterprise"},
							"source":          "custom",
						},
					},
					"exclude": []interface{}{
						map[string]interface{}{
							"name":            "status",
							"property_values": []interface{}{"inactive"},
						},
					},
				},
			},
		}

		input := []interface{}{originalMap}

		// Expand then flatten
		expanded := expandConditions(input, true)
		flattened := flattenConditions(expanded, true)

		if len(flattened) != 1 {
			t.Fatalf("Expected 1 flattened result after round-trip, got %d", len(flattened))
		}

		flattenedMap := flattened[0].(map[string]interface{})
		repoPropertySlice := flattenedMap["repository_property"].([]map[string]interface{})
		repoProperty := repoPropertySlice[0]

		includeSlice := repoProperty["include"].([]map[string]interface{})
		if includeSlice[0]["name"] != "tier" {
			t.Errorf("Round-trip failed: expected name 'tier', got '%v'", includeSlice[0]["name"])
		}
		if includeSlice[0]["source"] != "custom" {
			t.Errorf("Round-trip failed: expected source 'custom', got '%v'", includeSlice[0]["source"])
		}

		propertyValues := includeSlice[0]["property_values"].([]string)
		if len(propertyValues) != 2 || propertyValues[0] != "premium" || propertyValues[1] != "enterprise" {
			t.Errorf("Round-trip failed: property values mismatch, got %v", propertyValues)
		}
	})
}

func TestOrganizationPushRulesetSupport(t *testing.T) {
	// Test that organization push rulesets support all push-specific rules
	// This is a unit test since it only validates the expand/flatten functionality

	rulesMap := map[string]interface{}{
		"file_path_restriction": []interface{}{
			map[string]interface{}{
				"restricted_file_paths": []interface{}{"secrets/", "*.key", "private/"},
			},
		},
		"max_file_size": []interface{}{
			map[string]interface{}{
				"max_file_size": float64(10485760), // 10MB
			},
		},
		"max_file_path_length": []interface{}{
			map[string]interface{}{
				"max_file_path_length": 250,
			},
		},
		"file_extension_restriction": []interface{}{
			map[string]interface{}{
				"restricted_file_extensions": []interface{}{".exe", ".bat", ".sh", ".ps1"},
			},
		},
	}

	input := []interface{}{rulesMap}

	// Test expand functionality (organization rulesets use org=true)
	expandedRules := expandRules(input, true)

	if len(expandedRules) != 4 {
		t.Fatalf("Expected 4 expanded rules for organization push ruleset, got %d", len(expandedRules))
	}

	// Verify we have all expected push rule types
	ruleTypes := make(map[string]bool)
	for _, rule := range expandedRules {
		ruleTypes[rule.Type] = true
	}

	expectedPushRules := []string{"file_path_restriction", "max_file_size", "max_file_path_length", "file_extension_restriction"}
	for _, expectedType := range expectedPushRules {
		if !ruleTypes[expectedType] {
			t.Errorf("Expected organization push rule type %s not found in expanded rules", expectedType)
		}
	}

	// Test flatten functionality (organization rulesets use org=true)
	flattenedResult := flattenRules(expandedRules, true)

	if len(flattenedResult) != 1 {
		t.Fatalf("Expected 1 flattened result, got %d", len(flattenedResult))
	}

	flattenedRulesMap := flattenedResult[0].(map[string]interface{})

	// Verify file_path_restriction
	filePathRules := flattenedRulesMap["file_path_restriction"].([]map[string]interface{})
	if len(filePathRules) != 1 {
		t.Fatalf("Expected 1 file_path_restriction rule, got %d", len(filePathRules))
	}
	restrictedPaths := filePathRules[0]["restricted_file_paths"].([]string)
	if len(restrictedPaths) != 3 {
		t.Errorf("Expected 3 restricted file paths, got %d", len(restrictedPaths))
	}

	// Verify max_file_size
	maxFileSizeRules := flattenedRulesMap["max_file_size"].([]map[string]interface{})
	if len(maxFileSizeRules) != 1 {
		t.Fatalf("Expected 1 max_file_size rule, got %d", len(maxFileSizeRules))
	}
	if maxFileSizeRules[0]["max_file_size"] != int64(10485760) {
		t.Errorf("Expected max_file_size to be 10485760, got %v", maxFileSizeRules[0]["max_file_size"])
	}

	// Verify max_file_path_length
	maxFilePathLengthRules := flattenedRulesMap["max_file_path_length"].([]map[string]interface{})
	if len(maxFilePathLengthRules) != 1 {
		t.Fatalf("Expected 1 max_file_path_length rule, got %d", len(maxFilePathLengthRules))
	}
	if maxFilePathLengthRules[0]["max_file_path_length"] != 250 {
		t.Errorf("Expected max_file_path_length to be 250, got %v", maxFilePathLengthRules[0]["max_file_path_length"])
	}

	// Verify file_extension_restriction
	fileExtRules := flattenedRulesMap["file_extension_restriction"].([]map[string]interface{})
	if len(fileExtRules) != 1 {
		t.Fatalf("Expected 1 file_extension_restriction rule, got %d", len(fileExtRules))
	}
	restrictedExts := fileExtRules[0]["restricted_file_extensions"].([]string)
	if len(restrictedExts) != 4 {
		t.Errorf("Expected 4 restricted file extensions, got %d", len(restrictedExts))
	}
}
