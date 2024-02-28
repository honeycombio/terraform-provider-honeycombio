package provider

/*func Test_reconcileReadNotificationRecipientState(t *testing.T) {
	type args struct {
		remote []client.NotificationRecipient
		state  []models.NotificationRecipientModel
	}
	tests := []struct {
		name string
		args args
		want []models.NotificationRecipientModel
	}{
		{
			name: "both empty",
			args: args{},
			want: []models.NotificationRecipientModel{},
		},
		{
			name: "empty state",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "abcd12345", Type: client.RecipientTypeEmail, Target: "test@example.com"},
				},
				state: []models.NotificationRecipientModel{},
			},
			want: []models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345"), Type: types.StringValue("email"), Target: types.StringValue("test@example.com")},
			},
		},
		{
			name: "empty remote",
			args: args{
				remote: []client.NotificationRecipient{},
				state: []models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345"), Type: types.StringValue("email"), Target: types.StringValue("test@example.com")},
				},
			},
			want: []models.NotificationRecipientModel{},
		},
		{
			name: "remote and state reconciled",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "abcd12345", Type: client.RecipientTypeEmail, Target: "test@example.com"},
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-channel"},
				},
				state: []models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},                                           // defined by ID
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")}, // defined by type+target
				},
			},
			want: []models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345")},
				{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")},
			},
		},
		{
			name: "remote has additional recipients",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "abcd12345", Type: client.RecipientTypeEmail, Target: "test@example.com"},
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-channel"},
					{ID: "qrsty3847", Type: client.RecipientTypeSlack, Target: "#test-alerts"},
					{
						ID:     "ijkl13579",
						Type:   client.RecipientTypePagerDuty,
						Target: "test-pagerduty",
						Details: &client.NotificationRecipientDetails{
							PDSeverity: client.PDSeverityWARNING,
						}},
				},
				state: []models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},                                           // defined by ID
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")}, // defined by type+target
				},
			},
			want: []models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345")},
				{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")},
				{ID: types.StringValue("qrsty3847"), Type: types.StringValue("slack"), Target: types.StringValue("#test-alerts")},
				{
					ID:     types.StringValue("ijkl13579"),
					Type:   types.StringValue("pagerduty"),
					Target: types.StringValue("test-pagerduty"),
					Details: []models.NotificationRecipientDetailsModel{
						{PDSeverity: types.StringValue("warning")},
					},
				},
			},
		},
		{
			name: "state has additional recipients",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-foo"},
				},
				state: []models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-foo")},
					{ID: types.StringValue("ijkl13579"), Details: []models.NotificationRecipientDetailsModel{{PDSeverity: types.StringValue("warning")}}},
				},
			},
			want: []models.NotificationRecipientModel{
				{Type: types.StringValue("slack"), Target: types.StringValue("#test-foo")},
			},
		},
		{
			name: "state has totally unmatched recipients",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-foo"},
				},
				state: []models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")},
					{ID: types.StringValue("ijkl13579"), Details: []models.NotificationRecipientDetailsModel{{PDSeverity: types.StringValue("warning")}}},
				},
			},
			want: []models.NotificationRecipientModel{
				{ID: types.StringValue("efgh67890"), Type: types.StringValue("slack"), Target: types.StringValue("#test-foo")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, reconcileReadNotificationRecipientState(tt.args.remote, tt.args.state))
		})
	}
}*/
