package model

type ChunkTask struct {
	Name             string `yaml:"name"`
	Source           string `yaml:"source"`
	OutputDir        string `yaml:"output-dir"`
	SplitSize        int    `yaml:"split-size"`
	ConnectionMode   string `yaml:"connection_mode"`
	Miner            string `yaml:"miner"`
	DeltaURL         string `yaml:"delta-url"`
	DeltaToken       string `yaml:"delta-token"`
	DeltaWallet      string `yaml:"delta-wallet"`
	DeltaMetadataReq string `yaml:"delta-metadata-request"`
}

type Config struct {
	Label      string      `yaml:"label"`
	ChunkTasks []ChunkTask `yaml:"chunk-tasks"`
}

type Content struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	Size                 int64  `json:"size"`
	Cid                  string `json:"cid"`
	Miner                string `json:"miner"`
	ConnectionMode       string `json:"connection_mode"`
	Status               string `json:"status"`
	DeltaNodeUrl         string `json:"delta_node_url"`
	DeltaWallet          string `json:"delta_wallet"`
	DeltaMetadataRequest string `json:"delta_metadata_request"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

type ContentSplit struct {
	ID             int64  `json:"id"`
	ContentID      int64  `json:"content_id"`
	ContentCid     string `json:"split_cid"`
	Miner          string `json:"miner"`
	Status         int    `json:"status"`
	ConnectionMode string `json:"connection_mode"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// status
// 0 - queued
// 1 - chunking
// 2 - loading-to-delta
// 3 - success
// 4 - failed
