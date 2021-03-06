package main

import (
	"bufio"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unirita/cuto/testutil"
)

var testDataDir string

func TestMain(m *testing.M) {
	os.Exit(realTestMain(m))
}

func realTestMain(m *testing.M) int {
	testDataDir = filepath.Join(testutil.GetBaseDir(), "master", "_testdata")
	os.Chdir(testDataDir)
	os.RemoveAll("log")
	os.Mkdir("log", 0777)

	dbBase := filepath.Join(testDataDir, "data", "test.sqlite.org")
	dbFile := filepath.Join(testDataDir, "data", "test.sqlite")
	copyFile(dbBase, dbFile)
	defer os.RemoveAll(dbFile)

	return m.Run()
}

func copyFile(srcPath string, targetPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	target, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()

	r := bufio.NewReader(src)
	w := bufio.NewWriter(target)
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		w.Write(buf[:n])
	}
	return w.Flush()
}

func runTestServant(t *testing.T, waitInitCh chan<- struct{}) {
	listener, lerr := net.Listen("tcp", "localhost:15243")
	waitInitCh <- struct{}{}
	if lerr != nil {
		t.Log(lerr)
		return
	}

	conn, cerr := listener.Accept()
	if cerr != nil {
		t.Log(cerr)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	if _, err := conn.Read(buf); err != nil {
		t.Log(err)
		return
	}

	res := `{"type":"response","nid":1234,"jid":"j1","rc":0,"stat":1,"detail":"","var":"","st":"2015-04-01 12:34:56.789","et":"2015-04-01 12:35:46.123"}`
	res += "\n"

	if _, err := conn.Write([]byte(res)); err != nil {
		t.Log(err)
		return
	}
}

func TestFetchArgs_コマンドラインオプションを取得できる(t *testing.T) {
	os.Args = append(os.Args, "-v", "-n", "test", "-s", "-r", "123", "-c", "test.ini")
	args := fetchArgs()

	if args.versionFlag != flag_ON {
		t.Error("-vオプションの指定を検出できなかった。")
	}
	if args.networkName != "test" {
		t.Error("-nオプションの値を取得できなかった。")
	}
	if args.startFlag != flag_ON {
		t.Error("-sオプションの指定を検出できなかった。")
	}
	if args.rerunInstance != 123 {
		t.Error("-rオプションの指定を検出できなかった。")
	}
	if args.configPath != "test.ini" {
		t.Error("-cオプションの値を取得できなかった。")
	}
}

func TestRealMain_バージョン出力オプションが指定された場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.versionFlag = flag_ON

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_OK {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, Version) {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ネットワーク名およびインスタンスIDの両方が指定されなかった場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "INVALID ARGUMENT.") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ネットワーク名およびインスタンスIDの両方が指定された場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "test"
	args.rerunInstance = 123

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "EXCEPTION") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_存在しない設定ファイルが指定された場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "test"
	args.configPath = "noexists.ini"

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "FAILED TO READ EXPAND JOB CONFIG FILE") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_不正な内容の設定ファイルが指定された場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "test"
	args.configPath = "configerror.ini"

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "CONFIG PARM IS NOT EXACT FORMAT.") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_指定ネットワークの定義ファイルが存在しない(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "noexists"

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "FAILED TO READ BPMN FILE") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ログディレクトリが存在しない場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "test"
	args.configPath = "logerror.ini"

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "COULD NOT INITIALIZE LOGGER.") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ネットワーク定義の書式チェックのみを行う場合_エラーなし(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "test"

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_OK {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "IS VALID") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ネットワーク定義の書式チェックのみを行う場合_エラーあり(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "error"

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "IS NOT EXACT FORMAT") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ジョブ実行を行う_正常な実行(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "test"
	args.startFlag = flag_ON

	waitCh := make(chan struct{}, 1)
	go runTestServant(t, waitCh)
	<-waitCh

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_OK {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "STATUS [NORMAL]") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ジョブ実行を行う_拡張ジョブ定義にエラーあり(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "jobexerror"
	args.startFlag = flag_ON

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "FAILED TO READ EXPAND JOB CONFIG FILE") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_ジョブ実行を行う_ジョブ実行中にエラー発生(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.networkName = "runerror"
	args.startFlag = flag_ON

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "STATUS [ABNORMAL]") {
		t.Error("出力内容が想定と違っている。")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_リラン実行_既に正常終了済みの場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.rerunInstance = 1
	args.startFlag = flag_ON

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_OK {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "CTM029I") {
		t.Errorf("想定されるメッセージ[%s]が出力されていない。", "CTM029I")
		t.Logf("出力: %s", out)
	}
}

func TestRealMain_リラン実行_存在しないインスタンスIDの場合(t *testing.T) {
	c := testutil.NewStdoutCapturer()

	args := new(arguments)
	args.rerunInstance = 2
	args.startFlag = flag_ON

	c.Start()
	rc := realMain(args)
	out := c.Stop()

	if rc != rc_ERROR {
		t.Errorf("想定外のrc[%d]が返された。", rc)
	}
	if !strings.Contains(out, "CTM019E") {
		t.Errorf("想定されるメッセージ[%s]が出力されていない。", "CTM019I")
		t.Logf("出力: %s", out)
	}
}
