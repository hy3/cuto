// Copyright 2015 unirita Inc.
// Created 2015/04/10 shanxia

package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/unirita/cuto/console"
	"github.com/unirita/cuto/util"
)

const (
	defaultBindAddress       = `0.0.0.0`
	defaultBindPort          = 2015
	defaultHeartbeatSpanSec  = 30
	defaultMultiProc         = 20
	defaultDockerCommandPath = ``
	defaultDisuseJoblog      = 0
	defaultJobDir            = `jobscript`
	defaultJoblogDir         = `joblog`
	defaultLogDir            = `log`
	defaultOutputLevel       = `info`
	defaultMaxSizeKB         = 10240
	defaultMaxGeneration     = 2
	defaultTimeoutSec        = 1
)

const dirName = "bin"
const fileName = "servant.ini"
const tag_CUTOROOT = "<CUTOROOT>"

// 設定情報のデフォルト値を設定する。
func DefaultServantConfig() *ServantConfig {
	cfg := new(ServantConfig)
	cfg.Sys.BindAddress = defaultBindAddress
	cfg.Sys.BindPort = defaultBindPort
	cfg.Job.HeartbeatSpanSec = defaultHeartbeatSpanSec
	cfg.Job.MultiProc = defaultMultiProc
	cfg.Job.DockerCommandPath = defaultDockerCommandPath
	cfg.Job.DisuseJoblog = defaultDisuseJoblog
	cfg.Dir.JobDir = defaultJobDir
	cfg.Dir.JoblogDir = defaultJoblogDir
	cfg.Dir.LogDir = defaultLogDir
	cfg.Log.OutputLevel = defaultOutputLevel
	cfg.Log.MaxSizeKB = defaultMaxSizeKB
	cfg.Log.MaxGeneration = defaultMaxGeneration
	cfg.Log.TimeoutSec = defaultTimeoutSec

	return cfg

}

// サーバント設定情報
type ServantConfig struct {
	Sys sysSection
	Job jobSection
	Dir dirSection
	Log logSection
}

// サーバント設定のsysセクション
type sysSection struct {
	BindAddress string `toml:"bind_address"`
	BindPort    int    `toml:"bind_port"`
}

// サーバント設定のjobセクション
type jobSection struct {
	MultiProc         int    `toml:"multi_proc"`
	HeartbeatSpanSec  int    `toml:"heartbeat_span_sec"`
	DockerCommandPath string `toml:"docker_command_path"`
	DisuseJoblog      int    `toml:"disuse_joblog"`
}

// サーバント設定のdirセクション
type dirSection struct {
	JoblogDir string `toml:"joblog_dir"`
	JobDir    string `toml:"job_dir"`
	LogDir    string `toml:"log_dir"`
}

// 設定ファイルのlogセクション
type logSection struct {
	OutputLevel   string `toml:"output_level"`
	MaxSizeKB     int    `toml:"max_size_kb"`
	MaxGeneration int    `toml:"max_generation"`
	TimeoutSec    int    `toml:"timeout_sec"`
}

var Servant *ServantConfig
var FilePath string
var RootPath string

func init() {
	RootPath = util.GetRootPath()
	FilePath = fileName
}

// 設定ファイルを読み込む
// 読み込みに失敗する場合はDefaultServantConfig関数でデフォルト値を設定する。
//
// 戻り値: 設定値を格納したServantConfig構造体オブジェクト
func ReadConfig(configPath string) *ServantConfig {
	var err error
	if len(configPath) > 0 {
		FilePath = configPath
	}
	Servant, err = loadFile(FilePath)
	if err != nil {
		console.Display("CTS019E", err)
		console.Display("CTS004W", FilePath)
		Servant = DefaultServantConfig()
	}

	Servant.convertFullpath()
	return Servant
}

// 設定をリロードする。
//
// 戻り値: 設定値を格納したServantConfig構造体オブジェクト
func ReloadConfig() *ServantConfig {
	return ReadConfig(FilePath)
}

// 設定値のエラー検出を行う。
//
// 戻り値: エラー情報
func (c *ServantConfig) DetectError() error {
	if c.Sys.BindPort < 0 || 65535 < c.Sys.BindPort {
		return fmt.Errorf("sys.bind_port(%d) must be within the range 0 and 65535.", c.Sys.BindPort)
	}
	if c.Job.HeartbeatSpanSec <= 0 {
		return fmt.Errorf("job.heartbeat_span_sec(%d) must not be 0 or less.", c.Job.HeartbeatSpanSec)
	}
	if c.Job.MultiProc <= 0 {
		return fmt.Errorf("job.multi_proc(%d) must not be 0 or less.", c.Job.MultiProc)
	}
	if c.Log.MaxSizeKB <= 0 {
		return fmt.Errorf("log.max_size_kb(%d) must not be 0 or less.", c.Log.MaxSizeKB)
	}
	if c.Log.MaxGeneration <= 0 {
		return fmt.Errorf("log.max_generation(%d) must not be 0 or less.", c.Log.MaxGeneration)
	}

	return nil
}

func loadFile(filePath string) (*ServantConfig, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return loadReader(f)
}

func loadReader(reader io.Reader) (*ServantConfig, error) {
	sc := new(ServantConfig)
	if _, err := toml.DecodeReader(reader, sc); err != nil {
		return nil, err
	}

	sc.replaceCutoroot()

	return sc, nil
}

func (s *ServantConfig) replaceCutoroot() {
	s.Dir.JoblogDir = strings.Replace(s.Dir.JoblogDir, tag_CUTOROOT, util.GetRootPath(), -1)
	s.Dir.JobDir = strings.Replace(s.Dir.JobDir, tag_CUTOROOT, util.GetRootPath(), -1)
	s.Dir.LogDir = strings.Replace(s.Dir.LogDir, tag_CUTOROOT, util.GetRootPath(), -1)
}

// 設定値の相対パスを絶対パスへ変換する。
func (s *ServantConfig) convertFullpath() {
	if !filepath.IsAbs(s.Dir.JoblogDir) {
		s.Dir.JoblogDir = filepath.Join(RootPath, s.Dir.JoblogDir)
	}
	if !filepath.IsAbs(s.Dir.JobDir) {
		s.Dir.JobDir = filepath.Join(RootPath, s.Dir.JobDir)
	}
	if !filepath.IsAbs(s.Dir.LogDir) {
		s.Dir.LogDir = filepath.Join(RootPath, s.Dir.LogDir)
	}
}
