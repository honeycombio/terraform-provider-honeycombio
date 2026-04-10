package features

// DefaultFeatures returns the default features for the provider.
func DefaultFeatures() *Features {
	return &Features{
		Column:       defaultColumnFeatures(),
		Dataset:      defaultDatasetFeatures(),
		Intelligence: defaultIntelligenceFeatures(),
	}
}

func defaultColumnFeatures() FeaturesColumn {
	return FeaturesColumn{
		ImportOnConflict: false,
	}
}

func defaultDatasetFeatures() FeaturesDataset {
	return FeaturesDataset{
		ImportOnConflict: false,
	}
}

func defaultIntelligenceFeatures() FeaturesIntelligence {
	return FeaturesIntelligence{
		Enabled: false,
	}
}
