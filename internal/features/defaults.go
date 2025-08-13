package features

// DefaultFeatures returns the default features for the provider.
func DefaultFeatures() *Features {
	return &Features{
		Column: defaultColumnFeatures(),
	}
}

func defaultColumnFeatures() FeaturesColumn {
	return FeaturesColumn{
		ImportOnConflict: false,
	}
}
