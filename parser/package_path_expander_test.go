package parser

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
)

func incCnt(cnt *int) int {
	*cnt++
	return *cnt
}

func Test_packagePathExpander_calcLevel(t *testing.T) {
	var (
		cnt          int
		pathExpander = &packagePathExpander{}
	)

	type args struct {
		relPath string
	}
	tests := []struct {
		name      string
		p         *packagePathExpander
		args      args
		wantLevel int
	}{
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"path/"}, 1},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"/path/"}, 1},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"/path"}, 1},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"path"}, 1},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"p"}, 1},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"path/one"}, 2},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"path/one/"}, 2},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"/path/one"}, 2},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), pathExpander, args{"/path/one/"}, 2},
	}
	for _, tt := range tests {
		if gotLevel := tt.p.calcLevel(tt.args.relPath); gotLevel != tt.wantLevel {
			t.Errorf("%q. packagePathExpander.calcLevel() = %v, want %v", tt.name, gotLevel, tt.wantLevel)
		}
	}
}

func createFSTree(t *testing.T, fs afero.Fs, paths []string) {
	for _, path := range paths {
		if err := fs.MkdirAll(path, 0700); err != nil {
			t.Fatal(err.Error())
		}
	}
}

func Test_packagePathExpander_WalkDir(t *testing.T) {
	var (
		cnt    int
		baseFS = afero.NewMemMapFs()
	)

	createFSTree(t, baseFS, []string{
		"d1/d11/d111",
		"d2",
		"d2/d21/d121",
	})

	type args struct {
		baseFS afero.Fs
	}
	tests := []struct {
		name                     string
		p                        *packagePathExpander
		args                     args
		wantErr                  bool
		wantRelativePackagePaths []string
	}{
		{fmt.Sprintf("Case %d", incCnt(&cnt)), &packagePathExpander{0, nil}, args{baseFS}, false, []string{""}},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), &packagePathExpander{1, nil}, args{baseFS}, false, []string{"", "d1", "d2"}},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), &packagePathExpander{2, nil}, args{baseFS}, false, []string{"", "d1", "d1/d11", "d2", "d2/d21"}},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), &packagePathExpander{3, nil}, args{baseFS}, false, []string{"", "d1", "d1/d11", "d1/d11/d111", "d2", "d2/d21", "d2/d21/d121"}},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), &packagePathExpander{4, nil}, args{baseFS}, false, []string{"", "d1", "d1/d11", "d1/d11/d111", "d2", "d2/d21", "d2/d21/d121"}},
		{fmt.Sprintf("Case %d", incCnt(&cnt)), &packagePathExpander{MaxLevelInfinity, nil}, args{baseFS}, false, []string{"", "d1", "d1/d11", "d1/d11/d111", "d2", "d2/d21", "d2/d21/d121"}},
	}
	for _, tt := range tests {
		if err := tt.p.WalkDir(tt.args.baseFS); (err != nil) != tt.wantErr {
			t.Errorf("%q. packagePathExpander.WalkDir() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		if len(tt.wantRelativePackagePaths) != len(tt.p.relativePackagePaths) {
			t.Errorf("%q. packagePathExpander.WalkDir() len(relativePackagePaths) = %v, want %v", tt.name, len(tt.p.relativePackagePaths), len(tt.wantRelativePackagePaths))
		}
		for _, wantRel := range tt.wantRelativePackagePaths {
			containsWant := false
			for _, rel := range tt.p.relativePackagePaths {
				if rel == wantRel {
					containsWant = true
					break
				}
			}
			if !containsWant {
				t.Errorf("%q. packagePathExpander.WalkDir() relativePackagePaths does not contain %v. Only contained: %v", tt.name, wantRel, tt.p.relativePackagePaths)
			}
		}
	}
}
