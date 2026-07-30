package features

// DefaultFeatures returns the default features for the provider.
func DefaultFeatures() *Features {
	return &Features{
		Client:       defaultClientFeatures(),
		Column:       defaultColumnFeatures(),
		Dataset:      defaultDatasetFeatures(),
		Intelligence: defaultIntelligenceFeatures(),
	}
}

func defaultClientFeatures() FeaturesClient {
	return FeaturesClient{
		ProactiveThrottling: false,
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
