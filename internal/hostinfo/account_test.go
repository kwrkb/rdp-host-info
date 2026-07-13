package hostinfo

import (
	"reflect"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name       string
		pcName     string
		data       AccountData
		wantType   AccountType
		wantValues []string // Candidates の Value（順序込み）
		wantNotes  int      // Notes の件数
	}{
		{
			name:   "azure ad with UPN",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-12-1-1111111111-2222222222-3333333333-4444444444",
				Domain:  "AzureAD", User: "yugo",
				UPN: "yugo@example.com",
			},
			wantType:   AccountAzureAD,
			wantValues: []string{`AzureAD\yugo@example.com`},
		},
		{
			name:   "azure ad without UPN falls back to account name with note",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-12-1-1111111111-2222222222-3333333333-4444444444",
				Domain:  "AzureAD", User: "yugo",
			},
			wantType:   AccountAzureAD,
			wantValues: []string{`AzureAD\yugo`},
			wantNotes:  1,
		},
		{
			name:   "domain joined",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-5-21-1111111111-2222222222-3333333333-1001",
				Domain:  "CORP", User: "yugo",
				JoinKnown: true, DomainJoined: true, JoinedDomain: "corp.example.com",
			},
			wantType:   AccountDomain,
			wantValues: []string{`CORP\yugo`},
		},
		{
			name:   "domain joined machine but local user falls through",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-5-21-1111111111-2222222222-3333333333-1001",
				Domain:  "OMEN16", User: "yugo",
				JoinKnown: true, DomainJoined: true, JoinedDomain: "corp.example.com",
				MSAChecked: true,
			},
			wantType:   AccountLocal,
			wantValues: []string{`OMEN16\yugo`},
		},
		{
			name:   "microsoft account single email",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-5-21-1111111111-2222222222-3333333333-1001",
				Domain:  "OMEN16", User: "yugo",
				MSAChecked: true, MSASubKeys: []string{"user@example.com"},
			},
			wantType:   AccountMicrosoft,
			wantValues: []string{`MicrosoftAccount\user@example.com`, `OMEN16\yugo`},
			wantNotes:  1,
		},
		{
			name:   "microsoft account filters non-email subkeys",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-5-21-1111111111-2222222222-3333333333-1001",
				Domain:  "OMEN16", User: "yugo",
				MSAChecked: true, MSASubKeys: []string{"SomeGUID", "a@example.com", "b@example.net"},
			},
			wantType: AccountMicrosoft,
			wantValues: []string{
				`MicrosoftAccount\a@example.com`,
				`MicrosoftAccount\b@example.net`,
				`OMEN16\yugo`,
			},
			wantNotes: 1,
		},
		{
			name:   "local account",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-5-21-1111111111-2222222222-3333333333-1001",
				Domain:  "OMEN16", User: "yugo",
				MSAChecked: true,
			},
			wantType:   AccountLocal,
			wantValues: []string{`OMEN16\yugo`},
		},
		{
			name:   "msa registry unreadable is unknown with note",
			pcName: "OMEN16",
			data: AccountData{
				UserSID: "S-1-5-21-1111111111-2222222222-3333333333-1001",
				Domain:  "OMEN16", User: "yugo",
			},
			wantType:   AccountUnknown,
			wantValues: []string{`OMEN16\yugo`},
			wantNotes:  1,
		},
		{
			name:       "empty data has no candidates",
			pcName:     "OMEN16",
			data:       AccountData{},
			wantType:   AccountUnknown,
			wantValues: nil,
		},
		{
			name:   "empty pc name falls back to domain",
			pcName: "",
			data: AccountData{
				UserSID: "S-1-5-21-1111111111-2222222222-3333333333-1001",
				Domain:  "OMEN16", User: "yugo",
				MSAChecked: true,
			},
			wantType:   AccountLocal,
			wantValues: []string{`OMEN16\yugo`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.pcName, tt.data)
			if got.AccountType != tt.wantType {
				t.Errorf("AccountType = %v, want %v", got.AccountType, tt.wantType)
			}
			var values []string
			for _, c := range got.Candidates {
				values = append(values, c.Value)
			}
			if !reflect.DeepEqual(values, tt.wantValues) {
				t.Errorf("Candidate values = %v, want %v", values, tt.wantValues)
			}
			if len(got.Notes) != tt.wantNotes {
				t.Errorf("Notes = %v, want %d notes", got.Notes, tt.wantNotes)
			}
		})
	}
}
