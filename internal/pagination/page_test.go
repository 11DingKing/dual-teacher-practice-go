package pagination

import "testing"

func TestParsePagination(t *testing.T) {
	cases := []struct {
		name, limit, offset string
		wantL, wantO        int
	}{
		{"defaults", "", "", 20, 0}, {"valid", "10", "5", 10, 5}, {"zero limit", "0", "0", 20, 0}, {"negative limit", "-2", "-1", 20, 0}, {"huge", "1000", "3", 20, 3}, {"bad", "abc", "xyz", 20, 0}, {"one", "1", "0", 1, 0}, {"maximum", "100", "99", 100, 99},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := Parse(tc.limit, tc.offset)
			if p.Limit != tc.wantL || p.Offset != tc.wantO {
				t.Fatalf("got %#v", p)
			}
		})
	}
}
