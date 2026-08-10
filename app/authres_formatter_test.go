package app

import (
	"errors"
	"testing"

	"git.sr.ht/~rjarry/aerc/lib/authres"
)

func TestFormatAuthInfoToString(t *testing.T) {
	tests := []struct {
		name     string
		auth     *authres.Details
		showInfo bool
		maxLen   int
		want     string
	}{
		{
			name:     "Nil Auth Details",
			auth:     nil,
			showInfo: false,
			maxLen:   9999,
			want:     "(no header)",
		},
		{
			name: "Auth with Error",
			auth: &authres.Details{
				Err: errors.New("failed to parse header"),
			},
			showInfo: false,
			maxLen:   9999,
			want:     "failed to parse header",
		},
		{
			name: "Pass and Fail Results (Basic)",
			auth: &authres.Details{
				Results: []authres.Result{
					authres.ResultPass,
					authres.ResultFail,
					authres.ResultNeutral,
				},
			},
			showInfo: false,
			maxLen:   9999,
			want:     "✓ ✗ neutral",
		},
		{
			name: "With ShowInfo (Unlimited Length)",
			auth: &authres.Details{
				Results: []authres.Result{
					authres.ResultPass,
					authres.ResultPass,
				},
				Infos:   []string{"dkim=pass header.i=@example.com", "spf=pass"},
				Reasons: []string{" (ok)", ""},
			},
			showInfo: true,
			maxLen:   9999,
			want:     "✓ ✓ (dkim=pass header.i=@example.com (ok),spf=pass)",
		},
		{
			name: "With ShowInfo and Truncation (Max Length Applied)",
			auth: &authres.Details{
				Results: []authres.Result{
					authres.ResultPass,
				},
				Infos:   []string{"dkim=pass header.i=@very-long-subdomain.example.com"},
				Reasons: []string{" (ok)", ""},
			},
			showInfo: true,
			maxLen:   25,
			want:     "✓ (dkim=pass header.i=@…)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := FormatAuthInfoToChunks(tt.auth, tt.showInfo, tt.maxLen)
			var res string
			if err == nil {
				res = chunksToString(chunks)
			} else {
				res = err.Error()
			}
			if res != tt.want {
				t.Errorf("FormatAuthInfoToString()\n Got:  %q\n Want: %q", res, tt.want)
			}
		})
	}
}
