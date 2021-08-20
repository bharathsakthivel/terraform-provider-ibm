// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	// REMOVE below lines 13,14
	// "github.com/IBM-Cloud/bluemix-go/api/iamuum/iamuumv2"
	// "github.com/IBM-Cloud/bluemix-go/bmxerror"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"
)

func resourceIBMIAMDynamicRule() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMIAMDynamicRuleCreate,
		Read:     resourceIBMIAMDynamicRuleRead,
		Update:   resourceIBMIAMDynamicRuleUpdate,
		Delete:   resourceIBMIAMDynamicRuleDelete,
		Exists:   resourceIBMIAMDynamicRuleExists,
		Importer: &schema.ResourceImporter{},

		Schema: map[string]*schema.Schema{
			"access_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique identifier of the access group",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Rule",
			},
			"expiration": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The expiration in hours",
				ValidateFunc: validatePortRange(1, 24),
			},
			"identity_provider": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The realm name or identity proivider url",
			},
			"conditions": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "conditions info",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"claim": {
							Type:     schema.TypeString,
							Required: true,
						},
						"operator": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateAllowedStringValue([]string{"EQUALS", "EQUALS_IGNORE_CASE", "IN", "NOT_EQUALS_IGNORE_CASE", "NOT_EQUALS", "CONTAINS"}),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"rule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "id of the rule",
			},
		},
	}
}

func resourceIBMIAMDynamicRuleCreate(d *schema.ResourceData, meta interface{}) error {
	// REMOVE below line (83)
	// iamuumClient, err := meta.(ClientSession).IAMUUMAPIV2()
	iamAccessGroupsClient, err := meta.(ClientSession).IAMAccessGroupsV2()
	if err != nil {
		return err
	}

	grpID := d.Get("access_group_id").(string)
	name := d.Get("name").(string)
	realm := d.Get("identity_provider").(string)
	expiration := d.Get("expiration").(int)

	var cond []interface{}
	// REMOVE below line 97
	// condition := []iamuumv2.Condition{}
	// iamuumv2 => iamaccessgroupsv2
	condition := []iamaccessgroupsv2.Condition{}
	if res, ok := d.GetOk("conditions"); ok {
		cond = res.([]interface{})
		for _, e := range cond {
			r, _ := e.(map[string]interface{})
			// REMOVE below line 106
			// conditionParam := iamuumv2.Condition{
			conditionParam := iamaccessgroupsv2.Condition{
				Claim:    r["claim"].(string),
				Operator: r["operator"].(string),
				Value:    fmt.Sprintf("\"%s\"", r["value"].(string)),
			}
			condition = append(condition, conditionParam)
		}
	}

	// REMOVE below lin 117
	// createRuleReq := iamuumv2.CreateRuleRequest{
	createRuleReq := iamaccessgroupsv2.AddAccessGroupRuleule{
		Name:       name,
		RealmName:  realm,
		Expiration: expiration,
		Conditions: condition,
	}

	// REMOVE below line 126
	// response, err := iamuumClient.DynamicRule().Create(grpID, createRuleReq)
	response, err := iamAccessGroupsClient.DynamicRule().Create(grpID, createRuleReq)
	if err != nil {
		return err
	}
	ruleID := response.RuleID
	d.SetId(fmt.Sprintf("%s/%s", grpID, ruleID))

	return resourceIBMIAMDynamicRuleRead(d, meta)
}

func resourceIBMIAMDynamicRuleRead(d *schema.ResourceData, meta interface{}) error {
	// REMOVE below line 139
	// iamuumClient, err := meta.(ClientSession).IAMUUMAPIV2()
	iamAccessGroupsClient, err := meta.(ClientSession).IAMAccessGroupsV2()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	grpID := parts[0]
	ruleID := parts[1]

	// REMOVE below line 155
	// rules, _, err := iamuumClient.DynamicRule().Get(grpID, ruleID)
	rules, _, err := iamAccessGroupsClient.DynamicRule().Get(grpID, ruleID)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("Error retrieving access group Rules: %s", err)
	} else if err != nil && strings.Contains(err.Error(), "404") {
		d.SetId("")

		return nil
	}

	d.Set("access_group_id", grpID)
	d.Set("name", rules.Name)
	d.Set("expiration", rules.Expiration)
	d.Set("identity_provider", rules.RealmName)
	d.Set("conditions", flattenConditions(rules.Conditions))
	d.Set("rule_id", rules.RuleID)

	return nil
}

func resourceIBMIAMDynamicRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	// REMOVE below line 177
	// iamuumClient, err := meta.(ClientSession).IAMUUMAPIV2()
	iamAccessGroupsClient, err := meta.(ClientSession).IAMAccessGroupsV2()
	if err != nil {
		return err
	}
	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	grpID := parts[0]
	ruleID := parts[1]
	// REMOVE below line 191
	// _, etag, err := iamuumClient.DynamicRule().Get(grpID, ruleID)
	_, etag, err := iamAccessGroupsClient.DynamicRule().Get(grpID, ruleID)
	if err != nil {
		return fmt.Errorf("Error retrieving access group Rules: %s", err)
	}

	name := d.Get("name").(string)
	realm := d.Get("identity_provider").(string)
	expiration := d.Get("expiration").(int)

	var cond []interface{}
	// REMOVE below line 202
	// condition := []iamuumv2.Condition{}
	// iamuumv2 => iamaccessgroupsv2
	condition := []iamaccessgroupsv2.Condition{}
	if res, ok := d.GetOk("conditions"); ok {
		cond = res.([]interface{})
		for _, e := range cond {
			r, _ := e.(map[string]interface{})
			// REMOVE below line 211
			// conditionParam := iamuumv2.Condition{
			conditionParam := iamaccessgroupsv2.Condition{
				Claim:    r["claim"].(string),
				Operator: r["operator"].(string),
				Value:    fmt.Sprintf("\"%s\"", r["value"].(string)),
			}
			condition = append(condition, conditionParam)
		}
	}

	// REMOVE below lin 222
	// createRuleReq := iamuumv2.CreateRuleRequest{
	createRuleReq := iamaccessgroupsv2.AddAccessGroupRule{
		Name:       name,
		RealmName:  realm,
		Expiration: expiration,
		Conditions: condition,
	}
	// REMOVE below line 230
	// _, err = iamuumClient.DynamicRule().Replace(grpID, ruleID, createRuleReq, etag)
	_, err = iamAccessGroupsClient.DynamicRule().Replace(grpID, ruleID, createRuleReq, etag)
	if err != nil {
		return err
	}

	return resourceIBMIAMDynamicRuleRead(d, meta)

}

func resourceIBMIAMDynamicRuleDelete(d *schema.ResourceData, meta interface{}) error {
	// REMOVE below line 242
	// iamuumClient, err := meta.(ClientSession).IAMUUMAPIV2()
	iamAccessGroupsClient, err := meta.(ClientSession).IAMAccessGroupsV2()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	grpID := parts[0]
	ruleID := parts[1]

	// REMOVE below line 258
	// rules, _, err := iamuumClient.DynamicRule().Get(grpID, ruleID)
	rules, _, err := iamAccessGroupsClient.DynamicRule().Get(grpID, ruleID)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return err
	}

	d.SetId("")

	return nil
}

func resourceIBMIAMDynamicRuleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	// REMOVE below line 271
	// iamuumClient, err := meta.(ClientSession).IAMUUMAPIV2()
	iamAccessGroupsClient, err := meta.(ClientSession).IAMAccessGroupsV2()
	if err != nil {
		return false, err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return false, err
	}
	if len(parts) < 2 {
		return false, fmt.Errorf("Incorrect ID %s: Id should be a combination of accessGroupID/RuleID", d.Id())
	}
	grpID := parts[0]
	ruleID := parts[1]

	// REMOVE below line 289
	// rules, _, err := iamuumClient.DynamicRule().Get(grpID, ruleID)
	rules, _, err := iamAccessGroupsClient.DynamicRule().Get(grpID, ruleID)
	if err != nil {
		// CHECK
		if apiErr, ok := err.(bmxerror.RequestFailure); ok {
			if apiErr.StatusCode() == 404 {
				return false, nil
			}
		}
		return false, fmt.Errorf("Error communicating with the API: %s", err)
	}

	return rules.AccessGroupID == grpID, nil
}
