package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

// these tests are fragile but it's fine for now
func Test_fetcher(t *testing.T) {
	type args struct {
		tasks <-chan string
		wg    *sync.WaitGroup
	}
	type res struct {
		stdout string
		stderr string
	}

	tasksExample := make(chan string, 1)
	tasksExample <- "example.com"
	close(tasksExample)

	tasksUnknown := make(chan string, 1)
	tasksUnknown <- "unknown"
	close(tasksUnknown)

	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			"example.com test",
			args{
				tasksExample,
				&sync.WaitGroup{},
			},
			res{
				"http://example.com 84238dfc8092e5d9c0dac8ef93371a07\n",
				"",
			},
		},
		{
			"unknown test",
			args{
				tasksUnknown,
				&sync.WaitGroup{},
			},
			res{
				"",
				"Get \"http://unknown\": dial tcp: lookup unknown: no such host\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout // keep backup of the real stdout
			rStdout, wStdout, _ := os.Pipe()
			os.Stdout = wStdout
			oldStderr := os.Stderr // keep backup of the real stderr
			rStderr, wStderr, _ := os.Pipe()
			os.Stderr = wStderr

			tt.args.wg.Add(1)

			fetcher(tt.args.tasks, tt.args.wg)

			_ = wStderr.Close()
			_ = wStdout.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// check stdout
			var bufStdout bytes.Buffer
			_, err := io.Copy(&bufStdout, rStdout)
			if err != nil {
				t.Error(err)
			}
			if bufStdout.String() != tt.res.stdout {
				t.Errorf("Expected stdout: %s\ngot:%s\n", tt.res.stdout, bufStdout.String())
			}

			// check stderr
			var bufStderr bytes.Buffer
			_, err = io.Copy(&bufStderr, rStderr)
			if err != nil {
				t.Error(err)
			}
			if bufStderr.String() != tt.res.stderr {
				t.Errorf("Expected stderr: %s\ngot:%s\n", tt.res.stderr, bufStderr.String())
			}

		})
	}
}

func Test_run(t *testing.T) {
	type args struct {
		urls                []string
		maxParallelRequests int
	}
	type res struct {
		stdout map[string]struct{}
		stderr map[string]struct{}
	}

	tests := []struct {
		name string
		args args
		res  res
	}{
		{
			"simple success case",
			args{
				[]string{
					"example.com",
					"unknown.site",
					"https://account.habr.com/info/confidential/",
					"http://www.example.com/hello",
					"https://gist.githubusercontent.com/sooxl96/29bb90774fca2fb918adc9ad177e6609/raw/bcd587c19ed28f2b98bdc86dc2b70d8a7c61f6aa/terra-on-akash.yaml",
				},
				2,
			},
			res{
				map[string]struct{}{
					"http://example.com 84238dfc8092e5d9c0dac8ef93371a07":                          {},
					"https://account.habr.com/info/confidential/ 637e1307d9fbd4020139e67b325d3abe": {},
					"http://www.example.com/hello 84238dfc8092e5d9c0dac8ef93371a07":                {},
					"https://gist.githubusercontent.com/sooxl96/29bb90774fca2fb918adc9ad177e6609/raw/" +
						"bcd587c19ed28f2b98bdc86dc2b70d8a7c61f6aa/terra-on-akash.yaml 5780fa2aa770c6cc071f8070a84f4932": {},
				},
				map[string]struct{}{
					"Get \"http://unknown.site\": dial tcp: lookup unknown.site: no such host": {},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout // keep backup of the real stdout
			rStdout, wStdout, _ := os.Pipe()
			os.Stdout = wStdout
			oldStderr := os.Stderr // keep backup of the real stderr
			rStderr, wStderr, _ := os.Pipe()
			os.Stderr = wStderr

			run(tt.args.urls, tt.args.maxParallelRequests)

			_ = wStderr.Close()
			_ = wStdout.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// check stdout
			var bufStdout bytes.Buffer
			_, err := io.Copy(&bufStdout, rStdout)
			if err != nil {
				t.Error(err)
			}
			resStdout := strings.Split(strings.TrimSuffix(bufStdout.String(), "\n"), "\n")
			if len(resStdout) != len(tt.res.stdout) {
				t.Errorf("Expected stdout len: %d, got:%d\n", len(tt.res.stdout), len(resStdout))
			}
			for _, line := range resStdout {
				if _, ok := tt.res.stdout[line]; !ok {
					t.Errorf("Unexpected stdout: %s\n", line)
				}
			}

			// check stderr
			var bufStderr bytes.Buffer
			_, err = io.Copy(&bufStderr, rStderr)
			if err != nil {
				t.Error(err)
			}
			resStderr := strings.Split(strings.TrimSuffix(bufStderr.String(), "\n"), "\n")
			if len(resStderr) != len(tt.res.stderr) {
				t.Errorf("Expected stderr len: %d, got:%d\n", len(tt.res.stderr), len(resStderr))
			}
			for _, line := range resStderr {
				if _, ok := tt.res.stderr[line]; !ok {
					t.Errorf("Unexpected stderr: %s\n", line)
				}
			}
		})
	}
}
