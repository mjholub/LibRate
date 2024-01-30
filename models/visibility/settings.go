package visibility

type ObjectVisibilitySettings struct {
	LocallySearchable   bool `json:"locally_searchable,omitempty" db"locally_searchable" default:"true"`
	FederatedSearchable bool `json:"searchable_to_federated,omitempty" db:"searchable_to_federated" default:"true"`
}
