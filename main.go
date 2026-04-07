package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"evidence-bridge/internal/adb"
	"evidence-bridge/internal/engine"
	"evidence-bridge/internal/log"
	"evidence-bridge/internal/resource"
)

const configFileName = "config.ini"

type Config struct {
	StorageChoice string
}

func main() {
	// EXEのベースディレクトリ取得
	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)

	// 1. ログの初期化
	if err := log.InitLogger(baseDir); err != nil {
		fmt.Printf("ログ初期化失敗: %v\n", err)
		return
	}
	defer log.CloseLogger()

	// ログファイルを自動で開く (Windows)
	logPath := log.GetLogPath(baseDir)
	_ = exec.Command("cmd", "/c", "start", "notepad", logPath).Start()

	// UI表示 (クリーン)
	log.PrintUser("========================================")
	log.PrintUser("   Evidence-Bridge (爆速転送ツール)")
	log.PrintUser("========================================")

	// 2. 設定の読み込み
	cfg := loadConfig(baseDir)
	storagePath := "/sdcard"

	// 3. 環境構築 (内部ログは画面に出ない)
	adbExe, err := resource.ProvisionADB(baseDir)
	if err != nil {
		log.Error("ADB準備失敗: %v", err)
		log.PrintUser("[ERROR] システムの準備に失敗しました。ログを確認してください。")
		return
	}

	// 4. ADBクライアント初期化
	client, err := adb.NewClient(adbExe)
	if err != nil {
		log.Error("ADBクライアント初期化失敗: %v", err)
		log.PrintUser("[ERROR] ADB接続の初期化に失敗しました。ログを確認してください。")
		return
	}

	// 5. ストレージ設定
	if cfg.StorageChoice == "" {
		log.PrintUser("\n--- 初回設定: ストレージ選択 ---")
		log.PrintUser("1: 内部ストレージ (/sdcard)")
		log.PrintUser("2: SDカード (外部ストレージ)")
		fmt.Print("番号を入力してください (1 or 2): ")

		var choice string
		fmt.Scanln(&choice)
		if choice != "1" && choice != "2" {
			choice = "1"
		}
		cfg.StorageChoice = choice
		saveConfig(baseDir, cfg)
	}

	if cfg.StorageChoice == "2" {
		sdPath, err := client.GetSDCardPath()
		if err != nil {
			log.Error("SDカード検出失敗: %v", err)
			log.PrintUser("警告: SDカードが見つかりません。内部ストレージを使用します。")
			storagePath = "/sdcard"
		} else {
			storagePath = sdPath
		}
	}
	log.PrintUser("\n[設定済み] 保存先: %s", storagePath)

	sendDir := filepath.Join(baseDir, "Send")
	_ = os.MkdirAll(sendDir, 0755)

	eng := engine.NewEngine(client, nil, sendDir, storagePath)

	// 6. メインループ
	log.PrintUser("\n----------------------------------------")
	log.PrintUser(" 操作ガイド:")
	log.PrintUser(" [S] キー + Enter: Send フォルダ内のファイルを一括転送")
	log.PrintUser(" [Q] キー + Enter: 終了")
	log.PrintUser("----------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n待機中 (S=転送 / Q=終了) > ")
		if !scanner.Scan() {
			break
		}
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))

		switch input {
		case "s":
			if err := eng.TransferBatch(); err != nil {
				log.Error("バッチ転送エラー: %v", err)
			}
		case "q":
			log.PrintUser("終了します。")
			return
		}
	}
}

func loadConfig(baseDir string) Config {
	path := filepath.Join(baseDir, configFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	cfg := Config{}
	for _, line := range lines {
		if strings.HasPrefix(line, "StorageChoice=") {
			cfg.StorageChoice = strings.TrimPrefix(line, "StorageChoice=")
			cfg.StorageChoice = strings.TrimSpace(cfg.StorageChoice)
		}
	}
	return cfg
}

func saveConfig(baseDir string, cfg Config) {
	path := filepath.Join(baseDir, configFileName)
	content := fmt.Sprintf("StorageChoice=%s\n", cfg.StorageChoice)
	_ = os.WriteFile(path, []byte(content), 0666)
}
