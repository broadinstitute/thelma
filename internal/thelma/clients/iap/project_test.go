package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseProject(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Project
		wantErr bool
	}{
		{"DspDevopsSuperProd", args{"dsp-devops-super-prod"}, DspDevopsSuperProd, false},
		{"DspToolsK8s", args{"dsp-tools-k8s"}, DspToolsK8s, false},
		{"unknown", args{"unknown"}, -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProject(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseProject() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_ProjectID(t *testing.T) {
	tests := []struct {
		name    string
		p       Project
		want    string
		wantErr bool
	}{
		{"DspDevopsSuperProd", DspDevopsSuperProd, "dsp-devops-super-prod", false},
		{"DspToolsK8s", DspToolsK8s, "dsp-tools-k8s", false},
		{"unknown", -1, "unknown", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.ProjectID()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProjectID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProject_oauthCredentials(t *testing.T) {
	thelmaConfig, err := config.NewTestConfig(t, map[string]interface{}{
		"iap.projects.dspdevopssuperprod.clientid":     "super prod client id",
		"iap.projects.dspdevopssuperprod.clientsecret": "super prod client secret",
		"iap.projects.dsptoolsk8s.clientid":            "tools client id",
		"iap.projects.dsptoolsk8s.clientsecret":        "tools client secret",
	})
	require.NoError(t, err)
	var cfg iapConfig
	err = thelmaConfig.Unmarshal(configKey, &cfg)
	require.NoError(t, err)
	tests := []struct {
		name             string
		p                Project
		wantClientID     string
		wantClientSecret string
		wantErr          bool
	}{
		{
			name:             "DspDevopsSuperProd",
			p:                DspDevopsSuperProd,
			wantClientID:     "super prod client id",
			wantClientSecret: "super prod client secret",
			wantErr:          false,
		},
		{
			name:             "DspToolsK8s",
			p:                DspToolsK8s,
			wantClientID:     "tools client id",
			wantClientSecret: "tools client secret",
			wantErr:          false,
		},
		{
			name:    "unknown",
			p:       -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClientID, gotClientSecret, err := tt.p.oauthCredentials(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("oauthCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotClientID != tt.wantClientID {
				t.Errorf("oauthCredentials() gotClientID = %v, want %v", gotClientID, tt.wantClientID)
			}
			if gotClientSecret != tt.wantClientSecret {
				t.Errorf("oauthCredentials() gotClientSecret = %v, want %v", gotClientSecret, tt.wantClientSecret)
			}
		})
	}
}

func TestProject_tokenKey(t *testing.T) {
	tests := []struct {
		name    string
		p       Project
		want    string
		wantErr bool
	}{
		{"DspDevopsSuperProd", DspDevopsSuperProd, "dsp-devops-super-prod-iap-oauth-token", false},
		{"DspToolsK8s", DspToolsK8s, "dsp-tools-k8s-iap-oauth-token", false},
		{"unknown", -1, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.tokenKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("tokenKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
