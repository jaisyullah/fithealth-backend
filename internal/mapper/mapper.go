package mapper

import (
	"encoding/json"
	"fmt"
	"time"
)

// ToFHIRObservation maps minimal required fields to FHIR R4 Observation JSON
func ToFHIRObservation(patientRef string, obsType string, value float64, unit string, observedAt time.Time) ([]byte, error) {
	codes := map[string]map[string]string{
		"heart_rate": {"system": "http://loinc.org", "code": "8867-4", "display": "Heart rate"},
		"spo2":       {"system": "http://loinc.org", "code": "2708-6", "display": "Oxygen saturation in Arterial blood"},
	}

	c, ok := codes[obsType]
	if !ok {
		return nil, fmt.Errorf("unknown obs type %s", obsType)
	}

	fhir := map[string]interface{}{
		"resourceType": "Observation",
		"status":       "final",
		"category": []map[string]interface{}{
			{"coding": []map[string]string{{"system": "http://terminology.hl7.org/CodeSystem/observation-category", "code": "vital-signs"}}},
		},
		"code": map[string]interface{}{
			"coding": []map[string]string{{"system": c["system"], "code": c["code"], "display": c["display"]}},
			"text":   c["display"],
		},
		"subject":           map[string]string{"reference": fmt.Sprintf("Patient/%s", patientRef)},
		"effectiveDateTime": observedAt.UTC().Format(time.RFC3339),
		"valueQuantity": map[string]interface{}{
			"value":  value,
			"unit":   unit,
			"system": "http://unitsofmeasure.org",
			"code":   unit,
		},
	}
	return json.Marshal(fhir)
}
