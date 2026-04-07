package adb

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	adb "github.com/zach-klippenstein/goadb"
)

type Client struct {
	adbClient *adb.Adb
	adbPath   string
}

// NewClient は、既存のADBプロセスを終了させた後、指定されたパスのADBを使用して初期化します。
func NewClient(adbPath string) (*Client, error) {
	// 1. 既存のADBプロセスを強制終了 (ポートの重複やゾンビプロセス防止)
	killAdb()

	// 2. 指定された絶対パスを使用してサーバーを開始
	// goadb内部でも開始されますが、ここでは明示的にパスを指定して確実に起動させます。
	// また、サーバーが準備完了するまでリトライを行います。
	var lastErr error
	for i := 0; i < 5; i++ {
		cmd := exec.Command(adbPath, "start-server")
		if err := cmd.Run(); err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}

		// サーバーが応答するか確認
		config := adb.ServerConfig{
			PathToAdb: adbPath,
			Host:      "127.0.0.1",
			Port:      5037,
		}
		client, err := adb.NewWithConfig(config)
		if err == nil {
			if _, err := client.ServerVersion(); err == nil {
				return &Client{
					adbClient: client,
					adbPath:   adbPath,
				}, nil
			}
			lastErr = err
		} else {
			lastErr = err
		}

		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("[E002] ADBサーバーの初期化に失敗しました (5回試行): %v", lastErr)
}

// 動作中の adb.exe を強制終了します。
func killAdb() {
	// Windowsのタスクキルを使用。エラーは無視（プロセスがなくても正常終了扱いにするため）
	_ = exec.Command("taskkill", "/f", "/im", "adb.exe").Run()
	// 少し待機してOSがポートを解放するのを待つ
	time.Sleep(500 * time.Millisecond)
}

// GetAvailableDevice は、接続されている最初のオンライン端末を自動選択します。
func (c *Client) GetAvailableDevice() (*adb.Device, error) {
	devices, err := c.adbClient.ListDevices()
	if err != nil {
		return nil, fmt.Errorf("端末リスト取得失敗: %w", err)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("[E001] 端末が認識されていません：ケーブルを差し直すか、USBデバッグがONか確認してください")
	}

	// オンラインの端末を探す
	var selectedSerial string
	for _, devInfo := range devices {
		device := c.adbClient.Device(adb.DeviceWithSerial(devInfo.Serial))
		state, err := device.State()
		if err == nil && state == adb.StateOnline {
			selectedSerial = devInfo.Serial
			break
		}
	}

	if selectedSerial == "" {
		// 全台オフラインか未承認
		// 最初の1台の状態をチェックして詳細メッセージを出す
		device := c.adbClient.Device(adb.DeviceWithSerial(devices[0].Serial))
		state, _ := device.State()
		if state == adb.StateUnauthorized {
			return nil, fmt.Errorf("[E004] 端末が未承認です：Android画面で「USBデバッグを許可」してください")
		}
		return nil, fmt.Errorf("利用可能なオンライン端末が見つかりません (状態: %v)", state)
	}

	return c.adbClient.Device(adb.DeviceWithSerial(selectedSerial)), nil
}

// CheckFreeSpace は POSIX 形式 (-P) で容量をパースし、頑丈に空き容量をチェックします。
func (c *Client) CheckFreeSpace(device *adb.Device, path string) (int64, error) {
	out, err := device.RunCommand("df", "-P", path)
	if err != nil {
		return 0, fmt.Errorf("容量確認コマンド失敗: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(out))
	if !scanner.Scan() {
		return 0, fmt.Errorf("df 出力が空です")
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			var avail int64
			_, err := fmt.Sscanf(fields[3], "%d", &avail)
			if err == nil {
				return avail * 1024, nil
			}
		}
	}

	return 0, fmt.Errorf("df 出力のパースに失敗しました: %s", out)
}

func (c *Client) PushFile(device *adb.Device, localPath, remotePath string) (int64, error) {
	localFile, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("ファイル読み込み失敗: %w", err)
	}
	defer localFile.Close()

	stat, _ := localFile.Stat()

	_, _ = device.RunCommand("mkdir", "-p", filepath.ToSlash(filepath.Dir(remotePath)))

	remoteFile, err := device.OpenWrite(remotePath, 0644, time.Now())
	if err != nil {
		return 0, fmt.Errorf("Android側書き込みオープン失敗: %w", err)
	}
	defer remoteFile.Close()

	n, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return 0, fmt.Errorf("[E005] 転送中にエラーが発生しました：接続が切れた可能性があります (%w)", err)
	}

	if n != stat.Size() {
		return n, fmt.Errorf("転送サイズ不一致 (期待: %d, 実績: %d)", stat.Size(), n)
	}

	return n, nil
}

// GetSDCardPath はSDカードのパスを自動検出します
func (c *Client) GetSDCardPath() (string, error) {
	device, err := c.GetAvailableDevice()
	if err != nil {
		return "", err
	}

	// mountコマンドから外部SDカード（/storage/配下の英数字4桁-4桁）を探す
	out, err := device.RunCommand("mount")
	if err != nil {
		return "", err
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/storage/") {
			fields := strings.Fields(line)
			for _, f := range fields {
				if strings.HasPrefix(f, "/storage/") && !strings.Contains(f, "emulated") && !strings.Contains(f, "self") {
					return f, nil
				}
			}
		}
	}
	return "", fmt.Errorf("SDカードがマウントされていません")
}
