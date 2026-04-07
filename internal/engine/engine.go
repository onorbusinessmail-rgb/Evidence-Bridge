package engine

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"evidence-bridge/internal/adb"
	"evidence-bridge/internal/log"
	"evidence-bridge/internal/watcher"
)

type Engine struct {
	client      *adb.Client
	watcher     *watcher.Watcher
	sendDir     string
	storageBase string
	remoteBase  string
}

func NewEngine(client *adb.Client, watcher *watcher.Watcher, sendDir string, storageBase string) *Engine {
	return &Engine{
		client:      client,
		watcher:     watcher,
		sendDir:     sendDir,
		storageBase: storageBase,
		remoteBase:  "/Download/Evidence-Bridge/",
	}
}

func (e *Engine) TransferBatch() error {
	files, err := watcher.ScanExistingFiles(e.sendDir)
	if err != nil {
		return fmt.Errorf("ファイル一覧取得失敗: %w", err)
	}

	if len(files) == 0 {
		log.PrintUser("Send フォルダは空です。")
		return nil
	}

	log.PrintUser("%d 件のファイルを転送します...", len(files))
	successCount := 0
	for _, f := range files {
		if err := e.ProcessFile(f); err != nil {
			log.Error("転送失敗 [%s]: %v", filepath.Base(f), err)
			log.PrintUser("  [×] %s: 失敗 (%v)", filepath.Base(f), err)
		} else {
			successCount++
		}
	}

	log.PrintUser("完了: %d 件成功, %d 件失敗", successCount, len(files)-successCount)
	return nil
}

func (e *Engine) ProcessFile(localPath string) error {
	filename := filepath.Base(localPath)
	// Android側の絶対パスを構築
	fullRemoteDir := path.Join(e.storageBase, e.remoteBase)
	fullRemotePath := path.Join(fullRemoteDir, filename)

	// 1. Android接続チェック
	device, err := e.client.GetAvailableDevice()
	if err != nil {
		return err // E001, E004
	}

	// 2. 空き容量チェック
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("ローカルファイル情報取得失敗: %w", err)
	}

	// 容量取得
	free, err := e.client.CheckFreeSpace(device, fullRemoteDir)
	if err == nil {
		if free < info.Size() {
			return fmt.Errorf("[E003] Android側の容量不足 (必要: %d, 空き: %d)", info.Size(), free)
		}
	} else {
		log.Warn("空き容量の正確な取得に失敗しました。転送を強行します: %v", err)
	}

	// 3. 転送実行 & 計測
	log.Info("[%s] 転送開始...", filename)
	log.PrintUser("  [転送中] %s", filename)
	startTime := time.Now()

	bytesSent, err := e.client.PushFile(device, localPath, fullRemotePath)
	if err != nil {
		return err // E005
	}

	duration := time.Since(startTime)
	speedMBs := float64(bytesSent) / 1024 / 1024 / duration.Seconds()

	// 4. 検証成功後の削除
	if err := os.Remove(localPath); err != nil {
		log.Warn("[%s] 転送成功しましたが、元ファイルの削除に失敗しました: %v", filename, err)
	}

	log.Success("[%s] 移動完了 (%.2f MB/s) [サイズ: %d bytes, 時間: %v]",
		filename, speedMBs, bytesSent, duration.Truncate(time.Millisecond))

	return nil
}
