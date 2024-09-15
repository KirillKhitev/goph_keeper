package config

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"testing"
)

func ResetForTesting(usage func()) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	flag.CommandLine.Usage = flag.Usage
	flag.Usage = usage
}

func TestParamsClient_Parse(t *testing.T) {
	type fields struct {
		AddrServer string `json:"addr_server"`
	}

	tests := []struct {
		name         string
		fields       fields
		wantErr      bool
		createConfig bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				AddrServer: "127.0.0.1:8080",
			},
			wantErr:      false,
			createConfig: true,
		},
		{
			name: "positive test #2",
			fields: fields{
				AddrServer: "127.0.0.1:8080",
			},
			wantErr:      false,
			createConfig: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetForTesting(nil)
			var file *os.File

			if tt.createConfig {
				file, _ = os.Create(DefaultClientConfigPath)
				body, _ := json.Marshal(tt.fields)
				file.Write(body)
			}

			f := &ParamsClient{}
			if err := f.Parse(); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.createConfig {
				file.Close()
				err := os.RemoveAll(file.Name())
				if err != nil {
					t.Errorf("Failed to remove file: %v", err)
				}
			}
		})
	}
}

func TestParamsServer_Parse(t *testing.T) {
	type fields struct {
		AddrRun            string `json:"addr_run"`
		DBConnectionString string `json:"database_dsn"`
		MasterKey          string `json:"master_key"`
	}

	tests := []struct {
		name         string
		fields       fields
		wantErr      bool
		createConfig bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				AddrRun:            "127.0.0.1:8080",
				DBConnectionString: "db_connection_string",
				MasterKey:          "master_key",
			},
			wantErr:      false,
			createConfig: true,
		},
		{
			name: "negative test #2",
			fields: fields{
				AddrRun:            "127.0.0.1:8080",
				DBConnectionString: "db_connection_string",
			},
			wantErr:      true,
			createConfig: true,
		},
		{
			name: "negative test #3",
			fields: fields{
				AddrRun:   "127.0.0.1:8080",
				MasterKey: "master_key",
			},
			wantErr:      true,
			createConfig: true,
		},
		{
			name: "negative test #4",
			fields: fields{
				DBConnectionString: "db_connection_string",
				MasterKey:          "master_key",
			},
			wantErr:      true,
			createConfig: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetForTesting(nil)
			var file *os.File

			if tt.createConfig {
				file, _ = os.Create(DefaultServerConfigPath)
				body, _ := json.Marshal(tt.fields)
				file.Write(body)
			}

			f := &ParamsServer{}
			if err := f.Parse(); (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.createConfig {
				file.Close()
				err := os.RemoveAll(file.Name())
				if err != nil {
					t.Errorf("Failed to remove file: %v", err)
				}
			}
		})
	}
}
