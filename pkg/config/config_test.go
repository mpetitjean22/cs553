package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var validYAML = `Shard: 
  Name: shard0
  Index: 0
  Address: localhost:8080
  Replicas: [localhost:8083]
---
Shard: 
  Name: shard2
  Index: 2
  Address: localhost:8082
---
Shard:
  Name: shard1
  Index: 1
  Address: localhost:8081`

var correctConfig = Config{
	Shards: []Shard{
		{
			Shard: ShardConfig{
				Name:     "shard0",
				Index:    0,
				Address:  "localhost:8080",
				Replicas: []string{"localhost:8083"},
			},
		},
		{
			Shard: ShardConfig{
				Name:    "shard2",
				Index:   2,
				Address: "localhost:8082",
			},
		},
		{
			Shard: ShardConfig{
				Name:    "shard1",
				Index:   1,
				Address: "localhost:8081",
			},
		},
	},
	ShardToAddress: map[int]string{
		0: "localhost:8080",
		1: "localhost:8081",
		2: "localhost:8082",
	},
	TotalShards: 3,
}

func TestUnmarshalWithMultipleShards(t *testing.T) {
	yamlContents := []byte(validYAML)

	config := &Config{
		Shards: []Shard{},
	}
	err := config.unmarshalAllShards(yamlContents)
	if err != nil {
		t.Errorf("Unexpected Error Unmarshalling Shards: %v \n", err)
	}

	eq := reflect.DeepEqual(config.Shards, correctConfig.Shards)
	if !eq {
		t.Errorf("Unexpected result from unmarshalAllShards. Got %v Expected %v \n", config.Shards, correctConfig.Shards)
	}
}

func TestUnmarshalWithSingleShard(t *testing.T) {
	yamlContents := []byte(`Shard: 
  Name: shard0
  Index: 0
  Address: localhost:8080`)

	config := &Config{
		Shards: []Shard{},
	}
	err := config.unmarshalAllShards(yamlContents)
	if err != nil {
		t.Errorf("Unexpected Error Unmarshalling Shards: %v \n", err)
	}

	correctShards := []Shard{
		{
			ShardConfig{
				Name:    "shard0",
				Index:   0,
				Address: "localhost:8080",
			},
		},
	}
	eq := reflect.DeepEqual(config.Shards, correctShards)
	if !eq {
		t.Errorf("Unexpected result from unmarshalAllShards. Got %v Expected %v \n", config.Shards, correctShards)
	}
}

func TestUnmarshalWithBadYaml(t *testing.T) {
	yamlContents := []byte(`Shard: 
	NameBad: shard0
	Index: 0
	Address: localhost:8080`)

	config := &Config{
		Shards: []Shard{},
	}
	err := config.unmarshalAllShards(yamlContents)
	if err == nil {
		t.Errorf("Unexpected result, should error \n")
	}
}

func TestInvalidValidateNumShards(t *testing.T) {
	config := Config{
		Shards: []Shard{
			{
				Shard: ShardConfig{
					Name:  "shard0",
					Index: 0,
				},
			},
			{
				Shard: ShardConfig{
					Name:  "shard3",
					Index: 3,
				},
			},
			{
				Shard: ShardConfig{
					Name:  "shard4",
					Index: 4,
				},
			},
		},
	}
	valid := config.validateNumShards()
	if valid {
		t.Errorf("Expected valid to be false but got true \n")
	}
}

func TestValidValidateNumShards(t *testing.T) {
	valid := correctConfig.validateNumShards()
	if !valid {
		t.Errorf("Expected valid to be true but got false \n")
	}
}

func TestCreateShardToAddressMap(t *testing.T) {
	correctConfig.createShardToAddressMap()

	eq := reflect.DeepEqual(correctConfig.ShardToAddress, correctConfig.ShardToAddress)
	if !eq {
		t.Errorf("Error with shard to address mapping. Expected: %v Got: %v", correctConfig.ShardToAddress, correctConfig.ShardToAddress)
	}
}

func TestNewConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name())

	_, err = f.Write([]byte(validYAML))
	if err != nil {
		t.Error("Unexpected error with writing to the file: %w", err)
	}

	c, err := NewConfig(f.Name(), "shard0")
	if err != nil {
		t.Error("Unexpected error with NewConfig(): %w", err)
	}

	eq := reflect.DeepEqual(c.Shards, correctConfig.Shards)
	if !eq {
		t.Errorf("Incorrect Shards result. Expected: %v Got: %v \n", correctConfig.Shards, c.Shards)
	}

	eq = reflect.DeepEqual(c.ShardToAddress, correctConfig.ShardToAddress)
	if !eq {
		t.Errorf("Incorrect ShardToAddress result. Expected: %v Got: %v \n", correctConfig.ShardToAddress, c.ShardToAddress)
	}

	if c.TotalShards != correctConfig.TotalShards {
		t.Errorf("Incorrect TotalShards result. Expected: %v Got: %v \n", correctConfig.TotalShards, c.TotalShards)
	}

	if 0 != c.ShardIndex {
		t.Errorf("Incorrect ShardIndex result. Expected 0 Got: %d", c.ShardIndex)
	}

	badShardName, err := NewConfig(f.Name(), "notfound")
	if badShardName.ShardIndex != -1 {
		t.Errorf("Expected ShardIndex of shard not in config to be -1")
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}

func TestNewConfigBadFile(t *testing.T) {
	_, err := NewConfig("badFile.yaml", "shard0")
	if err == nil {
		t.Error("Expected error for a bad yaml file")
	}
}

func TestNewConfigWithBadYaml(t *testing.T) {
	badYamlContents := []byte(`Shard: 
	NameBad: shard0
	Index: 0
	Address: localhost:8080`)

	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name())

	_, err = f.Write([]byte(badYamlContents))
	if err != nil {
		t.Error("Unexpected error with writing to the file: %w", err)
	}

	_, err = NewConfig(f.Name(), "shard0")
	if err == nil {
		t.Error("Expected error for a unmarshaling bad yaml file")
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}

func TestNewConfigWithBadNumShards(t *testing.T) {
	badYamlContents := []byte(`Shard: 
  Name: shard2
  Index: 2
  Address: localhost:8080`)

	f, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Error("Unexpected error with opening the file: %w", err)
	}
	defer os.Remove(f.Name())

	_, err = f.Write([]byte(badYamlContents))
	if err != nil {
		t.Error("Unexpected error with writing to the file: %w", err)
	}

	_, err = NewConfig(f.Name(), "shard2")
	if err == nil {
		t.Error("Expected error for bad number of shards in yaml config")
	}

	err = f.Close()
	if err != nil {
		t.Error("Unexpected error with deleting the file: %w", err)
	}
}
