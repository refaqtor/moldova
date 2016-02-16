package moldova

import (
	"bytes"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestComparator func(string) error

type TestCase struct {
	Template     string
	Comparator   TestComparator
	ParseFailure bool
	WriteFailure bool
}

var GUIDCases = []TestCase{
	{
		Template: "{guid}",
		Comparator: func(s string) error {
			p := strings.Split(s, "-")
			if len(p) == 5 &&
				len(p[0]) == 8 &&
				len(p[1]) == len(p[2]) && len(p[2]) == len(p[3]) && len(p[3]) == 4 &&
				len(p[4]) == 12 {
				return nil
			}
			return errors.New("Guid not in correct format: " + s)
		},
	},
	{
		Template: "{guid}@{guid:ordinal:0}",
		Comparator: func(s string) error {
			p := strings.Split(s, "@")
			if p[0] == p[1] {
				return nil
			}
			return errors.New("Guid at position 1 not equal to guid at position 0 format: " + p[0] + " " + p[1])
		},
	},
	{
		Template:     "{guid}@{guid:ordinal:1}",
		WriteFailure: true,
	},
}

var NowCases = []TestCase{
	{
		// There is no proper deterministic way to test what the value of now is, without
		// something like rubys timecop (but the go-equivalent is not viable) or relying
		// on luck, which will run out if tests are run at just the wrong moment.
		// Therefore, for the basic test, i'm just asserting nothing went wrong for now.
		Template: "{now}",
		Comparator: func(s string) error {
			if len(s) > 0 {
				return nil
			}
			return errors.New("Guid not in correct format: " + s)
		},
	},
	{
		Template: "{now}@{now:ordinal:0}",
		Comparator: func(s string) error {
			p := strings.Split(s, "@")
			if p[0] == p[1] {
				return nil
			}
			return errors.New("Now at position 1 not equal to now at position 0 format: " + p[0] + " " + p[1])
		},
	},
	{
		Template:     "{now}@{now:ordinal:1}",
		WriteFailure: true,
	},
}

var AllCases = [][]TestCase{
	GUIDCases,
	NowCases,
}

// TODO Test each random function individually, under a number of inputs to make supported
// all the options behave as expected.

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(m.Run())
}

func TestAllCases(t *testing.T) {
	for _, cs := range AllCases {
		for _, c := range cs {
			cs, err := BuildCallstack(c.Template)
			// If we get an error and weren't expecting it
			// Or, if we didn't get one but were expecting it
			if err != nil && !c.ParseFailure {
				t.Error(err)
			} else if err == nil && c.ParseFailure {
				t.Error("Expected to encounter Parse Failure, but did not for Test Case ", c.Template)
			}

			result := &bytes.Buffer{}
			err = cs.Write(result)

			// If we get an error and weren't expecting it
			// Or, if we didn't get one but were expecting it
			if err != nil && !c.WriteFailure {
				t.Error(err)
			} else if err == nil && c.ParseFailure {
				t.Error("Expected to encounter Write Failure, but did not for Test Case ", c.Template)
			}

			if c.Comparator != nil {
				if err := c.Comparator(result.String()); err != nil {
					t.Error(err)
				}
			}
		}
	}
}

func TestBuildCallstack(t *testing.T) {
	template := "INSERT INTO floof VALUES ('{guid}','{guid:ordinal:0}','{country}',{int:min:-2000|max:0},{int:min:100|max:1000},{float:min:-1000.0|max:-540.0},{int:min:1|max:40},'{now}','{now:ordinal:0}','{unicode:length:2|case:up}',NULL,-3)"
	cs, err := BuildCallstack(template)
	if err != nil {
		t.Error(err)
	}
	result := &bytes.Buffer{}
	err = cs.Write(result)
	if err != nil {
		t.Error(err)
	}
}

func TestCountries(t *testing.T) {
	template := "INSERT INTO `floop` VALUES ('{country}','{country:case:up|ordinal:0}','{country}','{country:case:down|ordinal:1}')"
	cs, err := BuildCallstack(template)
	if err != nil {
		t.Error(err)
	}
	result := &bytes.Buffer{}
	err = cs.Write(result)
	if err != nil {
		t.Error(err)
	}
}

func TestInteger(t *testing.T) {
	template := "{int:min:5|max:6}"
	cs, err := BuildCallstack(template)
	if err != nil {
		t.Error(err)
	}
	result := &bytes.Buffer{}
	err = cs.Write(result)
	if err != nil {
		t.Error(err)
	}

	c, err := strconv.Atoi(result.String())
	if err != nil {
		t.Error(err)
	}
	if c < 5 || c > 6 {
		t.Error("Integer out of range")
	}
}

func TestTime(t *testing.T) {
	template := "{time:min:1|max:1|format:2006-01-02 15:04:05}"
	cs, err := BuildCallstack(template)
	if err != nil {
		t.Error(err)
	}
	result := &bytes.Buffer{}
	err = cs.Write(result)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkBuildCallstackRuns(b *testing.B) {
	template := "INSERT INTO floof VALUES ('{guid}','{time},'{guid:ordinal:0}','{country}',{int:min:-2000|max:0},{int:min:100|max:1000},{float:min:-1000.0|max:-540.0},{int:min:1|max:40},'{now}','{now:ordinal:0}','{unicode:length:2|case:up}',NULL,-3)"
	var cs *Callstack
	var err error
	for n := 0; n < b.N; n++ {
		if n == 0 {
			if cs, err = BuildCallstack(template); err != nil {
				b.Error(err)
			}
		}
		result := &bytes.Buffer{}
		err = cs.Write(result)
		if err != nil {
			b.Error(err)
		}
	}
}
