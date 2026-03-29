package shared

import "fmt"

// FieldProfile declares which optional fields are valid for a ResourceType category.
// Amount and ResourceType are always valid and omitted from the profile.
type FieldProfile struct {
	AllowTarget                     bool
	AllowSelectors                  bool
	AllowMaxTrigger                 bool
	AllowPer                        bool
	AllowTileRestrictions           bool
	AllowTileType                   bool
	AllowVariableAmount             bool
	AllowTemporary                  bool
	AllowOptional                   bool
	AllowPaymentAllowed             bool
	AllowTargetRestriction          bool
	AllowAllowDuplicatePlayerColony bool
}

// Category-level output profiles

var basicResourceOutputProfile = FieldProfile{
	AllowTarget:            true,
	AllowPer:               true,
	AllowVariableAmount:    true,
	AllowMaxTrigger:        true,
	AllowTargetRestriction: true,
}

var basicResourceInputProfile = FieldProfile{
	AllowTarget:         true,
	AllowVariableAmount: true,
	AllowOptional:       true,
}

var creditInputProfile = FieldProfile{
	AllowTarget:         true,
	AllowPaymentAllowed: true,
}

var cardStorageOutputProfile = FieldProfile{
	AllowTarget:    true,
	AllowSelectors: true,
	AllowPer:       true,
}

var cardStorageInputProfile = FieldProfile{
	AllowTarget:         true,
	AllowVariableAmount: true,
}

var productionOutputProfile = FieldProfile{
	AllowTarget:         true,
	AllowPer:            true,
	AllowVariableAmount: true,
}

var productionInputProfile = FieldProfile{
	AllowTarget: true,
}

var tilePlacementOutputProfile = FieldProfile{
	AllowTarget:           true,
	AllowTileRestrictions: true,
}

var genericTilePlacementOutputProfile = FieldProfile{
	AllowTarget:           true,
	AllowTileRestrictions: true,
	AllowTileType:         true,
}

var globalParameterOutputProfile = FieldProfile{
	AllowTarget: true,
}

var cardOperationOutputProfile = FieldProfile{
	AllowTarget:    true,
	AllowSelectors: true,
}

var discountOutputProfile = FieldProfile{
	AllowTarget:    true,
	AllowSelectors: true,
	AllowTemporary: true,
}

var paymentSubstituteOutputProfile = FieldProfile{
	AllowTarget:    true,
	AllowSelectors: true,
}

var storagePaymentSubstituteOutputProfile = FieldProfile{
	AllowSelectors: true,
}

var valueModifierOutputProfile = FieldProfile{
	AllowTarget:    true,
	AllowSelectors: true,
}

var effectOutputProfile = FieldProfile{
	AllowTarget:    true,
	AllowSelectors: true,
	AllowTemporary: true,
}

var colonyTileOutputProfile = FieldProfile{
	AllowTarget:                     true,
	AllowAllowDuplicatePlayerColony: true,
}

var tileReplacementOutputProfile = FieldProfile{
	AllowTarget:   true,
	AllowTileType: true,
}

var bonusTagsOutputProfile = FieldProfile{
	AllowTarget:    true,
	AllowPer:       true,
	AllowSelectors: true,
}

var targetOnlyOutputProfile = FieldProfile{
	AllowTarget: true,
}

var emptyProfile = FieldProfile{}

// outputFieldProfiles maps each ResourceType to its valid field profile for outputs.
var outputFieldProfiles = map[ResourceType]FieldProfile{
	// Basic resources
	ResourceCredit:   basicResourceOutputProfile,
	ResourceSteel:    basicResourceOutputProfile,
	ResourceTitanium: basicResourceOutputProfile,
	ResourcePlant:    basicResourceOutputProfile,
	ResourceEnergy:   basicResourceOutputProfile,
	ResourceHeat:     basicResourceOutputProfile,

	// Card storage resources
	ResourceMicrobe:      cardStorageOutputProfile,
	ResourceAnimal:       cardStorageOutputProfile,
	ResourceFloater:      cardStorageOutputProfile,
	ResourceScience:      cardStorageOutputProfile,
	ResourceAsteroid:     cardStorageOutputProfile,
	ResourceFighter:      cardStorageOutputProfile,
	ResourceDisease:      cardStorageOutputProfile,
	ResourceCardResource: cardStorageOutputProfile,

	// Card operations
	ResourceCardDraw:    cardOperationOutputProfile,
	ResourceCardTake:    cardOperationOutputProfile,
	ResourceCardPeek:    cardOperationOutputProfile,
	ResourceCardBuy:     cardOperationOutputProfile,
	ResourceCardDiscard: cardOperationOutputProfile,

	// Tile placements
	ResourceCityPlacement:     tilePlacementOutputProfile,
	ResourceOceanPlacement:    tilePlacementOutputProfile,
	ResourceGreeneryPlacement: tilePlacementOutputProfile,
	ResourceVolcanoPlacement:  tilePlacementOutputProfile,
	ResourceTilePlacement:     genericTilePlacementOutputProfile,

	// Tile count types (used in PerCondition, not directly as conditions)
	ResourceCityTile:            emptyProfile,
	ResourceOceanTile:           emptyProfile,
	ResourceGreeneryTile:        emptyProfile,
	ResourceVolcanoTile:         emptyProfile,
	ResourceNaturalPreserveTile: emptyProfile,
	ResourceMiningTile:          emptyProfile,
	ResourceNuclearZoneTile:     emptyProfile,
	ResourceEcologicalZoneTile:  emptyProfile,
	ResourceMoholeTile:          emptyProfile,
	ResourceRestrictedTile:      emptyProfile,
	ResourceLandTile:            emptyProfile,
	ResourceNonOceanTile:        emptyProfile,
	ResourceOceanSpace:          emptyProfile,

	// Colony
	ResourceColonyTile:  colonyTileOutputProfile,
	ResourceColonyCount: targetOnlyOutputProfile,
	ResourceColonyBonus: targetOnlyOutputProfile,

	// Global parameters
	ResourceTemperature:     globalParameterOutputProfile,
	ResourceOxygen:          globalParameterOutputProfile,
	ResourceOcean:           globalParameterOutputProfile,
	ResourceVenus:           globalParameterOutputProfile,
	ResourceTR:              {AllowTarget: true, AllowPer: true},
	ResourceGlobalParameter: globalParameterOutputProfile,

	// Production
	ResourceCreditProduction:   productionOutputProfile,
	ResourceSteelProduction:    productionOutputProfile,
	ResourceTitaniumProduction: productionOutputProfile,
	ResourcePlantProduction:    productionOutputProfile,
	ResourceEnergyProduction:   productionOutputProfile,
	ResourceHeatProduction:     productionOutputProfile,
	ResourceAnyProduction:      productionOutputProfile,

	// Effects
	ResourceEffect:                   effectOutputProfile,
	ResourceTag:                      effectOutputProfile,
	ResourceDiscount:                 discountOutputProfile,
	ResourcePaymentSubstitute:        paymentSubstituteOutputProfile,
	ResourceStoragePaymentSubstitute: storagePaymentSubstituteOutputProfile,
	ResourceValueModifier:            valueModifierOutputProfile,
	ResourceGlobalParameterLenience:  effectOutputProfile,
	ResourceIgnoreGlobalRequirements: effectOutputProfile,
	ResourceOceanAdjacencyBonus:      effectOutputProfile,
	ResourceDefense:                  effectOutputProfile,
	ResourceActionReuse:              targetOnlyOutputProfile,

	// Special
	ResourceLandClaim:       targetOnlyOutputProfile,
	ResourceExtraActions:    targetOnlyOutputProfile,
	ResourceTileDestruction: targetOnlyOutputProfile,
	ResourceTileReplacement: tileReplacementOutputProfile,
	ResourceBonusTags:       bonusTagsOutputProfile,
	ResourceWorldTreeTile:   targetOnlyOutputProfile,
	ResourceAwardFund:       targetOnlyOutputProfile,
	ResourceFreeTrade:       targetOnlyOutputProfile,
	ResourceCardCount:       emptyProfile,
	ResourceColonyTrackStep: targetOnlyOutputProfile,
}

// inputFieldProfiles maps each ResourceType to its valid field profile for inputs.
// Most resource types never appear as inputs. Types not in this map are invalid as inputs.
var inputFieldProfiles = map[ResourceType]FieldProfile{
	// Basic resources
	ResourceCredit:   creditInputProfile,
	ResourceSteel:    basicResourceInputProfile,
	ResourceTitanium: basicResourceInputProfile,
	ResourcePlant:    basicResourceInputProfile,
	ResourceEnergy:   basicResourceInputProfile,
	ResourceHeat:     basicResourceInputProfile,

	// Card storage resources (can be spent from cards)
	ResourceMicrobe:  cardStorageInputProfile,
	ResourceAnimal:   cardStorageInputProfile,
	ResourceFloater:  cardStorageInputProfile,
	ResourceScience:  cardStorageInputProfile,
	ResourceAsteroid: cardStorageInputProfile,
	ResourceFighter:  cardStorageInputProfile,

	// Card operations
	ResourceCardDiscard: {AllowTarget: true, AllowOptional: true, AllowVariableAmount: true},

	// Production (can be reduced)
	ResourceCreditProduction: productionInputProfile,
	ResourceEnergyProduction: productionInputProfile,
}

// ValidTargets is the set of known valid target values.
var validTargets = map[string]bool{
	"self-player":         true,
	"self-card":           true,
	"any-card":            true,
	"any-player":          true,
	"all-opponents":       true,
	"none":                true,
	"steal-any-player":    true,
	"steal-from-any-card": true,
	"":                    true,
}

// AllResourceTypes contains every ResourceType constant for exhaustiveness testing.
var AllResourceTypes = []ResourceType{
	ResourceCredit, ResourceSteel, ResourceTitanium, ResourcePlant, ResourceEnergy, ResourceHeat,
	ResourceMicrobe, ResourceAnimal, ResourceFloater, ResourceScience, ResourceAsteroid, ResourceFighter, ResourceDisease,
	ResourceCardDraw, ResourceCardTake, ResourceCardPeek, ResourceCardBuy, ResourceCardDiscard,
	ResourceCityPlacement, ResourceOceanPlacement, ResourceGreeneryPlacement, ResourceVolcanoPlacement, ResourceTilePlacement,
	ResourceCityTile, ResourceOceanTile, ResourceGreeneryTile, ResourceVolcanoTile,
	ResourceColonyTile, ResourceColonyCount, ResourceColonyBonus,
	ResourceNaturalPreserveTile, ResourceMiningTile, ResourceNuclearZoneTile, ResourceEcologicalZoneTile, ResourceMoholeTile, ResourceRestrictedTile,
	ResourceLandTile, ResourceNonOceanTile, ResourceOceanSpace,
	ResourceTemperature, ResourceOxygen, ResourceOcean, ResourceVenus, ResourceTR, ResourceGlobalParameter,
	ResourceCreditProduction, ResourceSteelProduction, ResourceTitaniumProduction, ResourcePlantProduction, ResourceEnergyProduction, ResourceHeatProduction, ResourceAnyProduction,
	ResourceEffect, ResourceTag,
	ResourceGlobalParameterLenience, ResourceIgnoreGlobalRequirements, ResourceDefense, ResourceDiscount, ResourceValueModifier, ResourcePaymentSubstitute, ResourceOceanAdjacencyBonus,
	ResourceLandClaim, ResourceStoragePaymentSubstitute, ResourceCardResource, ResourceActionReuse,
	ResourceExtraActions, ResourceTileDestruction, ResourceTileReplacement, ResourceBonusTags,
	ResourceWorldTreeTile, ResourceAwardFund, ResourceFreeTrade, ResourceCardCount,
	ResourceColonyTrackStep,
}

// ValidateResourceCondition checks that only valid fields are set for the given ResourceType.
// isInput indicates whether this is an input (true) or output (false).
// Returns a list of violation descriptions (empty = valid).
func ValidateResourceCondition(bc BehaviorCondition, isInput bool) []string {
	rc := flattenCondition(bc)
	var violations []string

	profiles := outputFieldProfiles
	direction := "output"
	if isInput {
		profiles = inputFieldProfiles
		direction = "input"
	}

	profile, ok := profiles[rc.ResourceType]
	if !ok {
		return []string{fmt.Sprintf("resource type %q is not valid as %s", rc.ResourceType, direction)}
	}

	if rc.Target != "" && rc.Target != "none" && !profile.AllowTarget {
		violations = append(violations, fmt.Sprintf("field 'target' (%q) not allowed for %s %q", rc.Target, direction, rc.ResourceType))
	}

	if rc.Target != "" && !validTargets[rc.Target] {
		violations = append(violations, fmt.Sprintf("unknown target value %q", rc.Target))
	}

	if len(rc.Selectors) > 0 && !profile.AllowSelectors {
		violations = append(violations, fmt.Sprintf("field 'selectors' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.MaxTrigger != nil && !profile.AllowMaxTrigger {
		violations = append(violations, fmt.Sprintf("field 'maxTrigger' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.Per != nil && !profile.AllowPer {
		violations = append(violations, fmt.Sprintf("field 'per' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.TileRestrictions != nil && !profile.AllowTileRestrictions {
		violations = append(violations, fmt.Sprintf("field 'tileRestrictions' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.TileType != "" && !profile.AllowTileType {
		violations = append(violations, fmt.Sprintf("field 'tileType' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.VariableAmount && !profile.AllowVariableAmount {
		violations = append(violations, fmt.Sprintf("field 'variableAmount' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.Temporary != "" && !profile.AllowTemporary {
		violations = append(violations, fmt.Sprintf("field 'temporary' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.Optional && !profile.AllowOptional {
		violations = append(violations, fmt.Sprintf("field 'optional' not allowed for %s %q", direction, rc.ResourceType))
	}

	if len(rc.PaymentAllowed) > 0 && !profile.AllowPaymentAllowed {
		violations = append(violations, fmt.Sprintf("field 'paymentAllowed' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.TargetRestriction != nil && !profile.AllowTargetRestriction {
		violations = append(violations, fmt.Sprintf("field 'targetRestriction' not allowed for %s %q", direction, rc.ResourceType))
	}

	if rc.AllowDuplicatePlayerColony && !profile.AllowAllowDuplicatePlayerColony {
		violations = append(violations, fmt.Sprintf("field 'allowDuplicatePlayerColony' not allowed for %s %q", direction, rc.ResourceType))
	}

	return violations
}

// GetOutputProfile returns the output FieldProfile for a ResourceType and whether it exists.
func GetOutputProfile(rt ResourceType) (FieldProfile, bool) {
	p, ok := outputFieldProfiles[rt]
	return p, ok
}

// GetInputProfile returns the input FieldProfile for a ResourceType and whether it exists.
func GetInputProfile(rt ResourceType) (FieldProfile, bool) {
	p, ok := inputFieldProfiles[rt]
	return p, ok
}
