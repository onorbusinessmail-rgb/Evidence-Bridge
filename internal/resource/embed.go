package resource

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"evidence-bridge/internal/log"
)

//go:embed bin/*
var adbFiles embed.FS

// ProvisionADB は、埋め込まれたADBバイナリを展開します。
func ProvisionADB(exeDir string) (string, error) {
	destDir := filepath.Join(exeDir, "bin")

	// 1. 書き込み権限テスト
	if err := testWrite(destDir); err != nil {
		log.Warn("カレントディレクトリへの書き込み制限を検知しました。Tempフォルダへ展開します。")
		destDir = filepath.Join(os.TempDir(), "Evidence-Bridge", "bin")
	}

	// ディレクトリ作成
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("ディレクトリ作成失敗: %w", err)
	}

	files := []string{"adb.exe", "AdbWinApi.dll", "AdbWinUsbApi.dll"}
	for _, filename := range files {
		destPath := filepath.Join(destDir, filename)
		
		// 既存バイナリの検証
		if isValid, err := validateBinary(destPath, filename == "adb.exe"); !isValid {
			if err != nil {
				log.Warn("[%s] 無効なバイナリを検知しました。再展開を試みます: %v", filename, err)
			}
			// 無効なら削除して再展開を促す
			_ = os.Remove(destPath)
		} else {
			// 有効ならスキップ
			continue
		}

		// 展開 (リトライ付き)
		if err := extractFileWithRetry(filename, destPath); err != nil {
			return "", err
		}

		// 展開後の検証（埋め込み自体が壊れている場合のエラー通知）
		if isValid, err := validateBinary(destPath, filename == "adb.exe"); !isValid {
			return "", fmt.Errorf("展開されたバイナリ [%s] が無効です：%v (埋め込みデータ自体が空か壊れている可能性があります)", filename, err)
		}
	}

	adbPath := filepath.Join(destDir, "adb.exe")
	log.Info("ADBバイナリ準備完了: %s", adbPath)
	return adbPath, nil
}

// validateBinary はファイルの有効性をチェックします。
func validateBinary(path string, checkMZ bool) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, nil // 存在しない
	}

	if info.Size() == 0 {
		return false, fmt.Errorf("ファイルサイズが 0 です")
	}

	if checkMZ {
		f, err := os.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()

		header := make([]byte, 2)
		_, err = f.Read(header)
		if err != nil {
			return false, err
		}
		// Windows実行ファイル (PE) は "MZ" で始まる
		if !bytes.Equal(header, []byte("MZ")) {
			return false, fmt.Errorf("Windows実行形式 (MZ) ではありません")
		}
	}

	return true, nil
}

func testWrite(dir string) error {
	_ = os.MkdirAll(dir, 0755)
	testFile := filepath.Join(dir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	f.Close()
	_ = os.Remove(testFile)
	return nil
}

func extractFileWithRetry(filename, destPath string) error {
	srcPath := "bin/" + filename
	
	var lastErr error
	for i := 0; i < 3; i++ {
		err := extract(srcPath, destPath)
		if err == nil {
			return nil
		}
		lastErr = err
		log.Warn("[%s] 展開リトライ中 (%d/3)... %v", filename, i+1, err)
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("ファイル [%s] の展開に失敗しました: %w", filename, lastErr)
}

func extract(srcPath, destPath string) error {
	src, err := adbFiles.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	
	// 強制書き込み
	return dst.Sync()
}
