package main

import (
	"fmt"
	"log"
	"web-api/back/models"
)

func main() {
	fmt.Println("Linux Audit API - domain model")

	host := models.Host{
		ID:       1,
		Name:     "Test Server",
		Hostname: "ubuntu-test",
		OS:       "Ubuntu Linux",
	}

	if err := host.Validate(); err != nil {
		log.Fatal(err)
	}

	policy := models.SecurityPolicy{
		ID:   1,
		Name: "Default Linux Policy",
	}

	rule := models.PolicyRule{
		ID:       1,
		Type:     models.RuleUserAdded,
		Severity: models.SeverityHigh,
		Enabled:  true,
	}

	if err := policy.AddRule(rule); err != nil {
		log.Fatal(err)
	}

	if err := host.AssignPolicy(policy.ID); err != nil {
		log.Fatal(err)
	}

	audit, err := models.NewAudit(host.ID)
	if err != nil {
		log.Fatal(err)
	}

	audit.LinuxUsers = []models.LinuxUser{
		{
			Username: "root",
			UID:      0,
			GID:      0,
		},
		{
			Username: "admin",
			UID:      1000,
			GID:      1000,
		},
	}

	if err := audit.Complete(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Host:", host.Hostname)
	fmt.Println("Policy:", policy.Name)
	fmt.Println("Active rules:", len(policy.ActiveRules()))
	fmt.Println("Audit status:", audit.Status)
	fmt.Println("Linux users:", len(audit.LinuxUsers))
	fmt.Println("Total objects:", audit.TotalObjects())
}
