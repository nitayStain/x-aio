package operations

/*
This type represents the GraphQL operation.
*/
type Operation struct {
	QueryID         string   `json:"queryId"`
	OperationName   string   `json:"operationName"`
	OperationType   string   `json:"operationType"`
	FeatureSwitches []string `json:"featureSwitches"`
	FieldToggles    []string `json:"fieldToggles"`
}
