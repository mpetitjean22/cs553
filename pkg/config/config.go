package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Shards         []Shard
	ShardToAddress map[int]string
	ShardIndex     int
	TotalShards    int
}

type Shard struct {
	Shard ShardConfig `yaml:"Shard"`
}

type ShardConfig struct {
	Name     string   `yaml:"Name"`
	Index    int      `yaml:"Index"`
	Address  string   `yaml:"Address"`
	Replicas []string `yaml:"Replicas"`
}

func (c *Config) unmarshalAllShards(yamlFile []byte) error {
	r := bytes.NewReader(yamlFile)
	decoder := yaml.NewDecoder(r)
	for {
		var shard Shard
		if err := decoder.Decode(&shard); err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		c.Shards = append(c.Shards, shard)
	}
	return nil
}

func (c *Config) validateNumShards() bool {
	maxIndex := -1
	for _, s := range c.Shards {
		if s.Shard.Index > maxIndex {
			maxIndex = s.Shard.Index
		}
	}
	return maxIndex+1 == len(c.Shards)
}

func (c *Config) createShardToAddressMap() {
	shardToAddress := make(map[int]string)
	for _, s := range c.Shards {
		shardToAddress[s.Shard.Index] = s.Shard.Address
	}
	c.ShardToAddress = shardToAddress
}

func NewConfig(fileName string, shardName string) (*Config, error) {
	config := &Config{
		Shards: []Shard{},
	}

	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("Could not find Config YAML: %s", fileName)
	}

	if err := config.unmarshalAllShards(yamlFile); err != nil {
		return nil, fmt.Errorf("could not parse yaml file: %w", err)
	}

	if valid := config.validateNumShards(); !valid {
		return nil, fmt.Errorf("shard index greater than number of shards")
	}

	config.TotalShards = len(config.Shards)
	config.ShardIndex = -1
	for _, s := range config.Shards {
		if s.Shard.Name == shardName {
			config.ShardIndex = s.Shard.Index
		}
	}
	config.createShardToAddressMap()

	return config, nil
}
