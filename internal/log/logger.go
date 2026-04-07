package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	fileWriter io.WriteCloser
	multiWriter io.Writer
)

// InitLogger はログ機能を初期化します。
// 実行ファイルと同じ階層に書き込めない場合は Temp フォルダに作成します。
func InitLogger(exeBaseDir string) error {
	logFileName := "bridge_log.txt"
	logPath := filepath.Join(exeBaseDir, logFileName)

	// 書き込みテスト
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// カレントディレクトリが書き込み不可の場合、Temp フォルダへ
		logPath = filepath.Join(os.TempDir(), "Evidence-Bridge", logFileName)
		_ = os.MkdirAll(filepath.Dir(logPath), 0755)
		f, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("ログファイルの作成に失敗しました: %w", err)
		}
	}

	fileWriter = f
	multiWriter = fileWriter
	
	// 標準ログの出力を MultiWriter に設定
	log.SetOutput(multiWriter)
	// プレフィックスなし（自前でタイムスタンプを入れるため）
	log.SetFlags(0)

	Info("Logging initialized. Path: %s", logPath)
	return nil
}

// CloseLogger はログファイルを閉じます。
func CloseLogger() {
	if fileWriter != nil {
		fileWriter.Close()
	}
}

func format(level, msg string) string {
	return fmt.Sprintf("[%s] [%s] %s", time.Now().Format("2006-01-02 15:04:05.000"), level, msg)
}

func Info(formatStr string, v ...interface{}) {
	log.Println(format( "INFO", fmt.Sprintf(formatStr, v...)))
}

func Success(formatStr string, v ...interface{}) {
	log.Println(format("SUCCESS", fmt.Sprintf(formatStr, v...)))
}

func Warn(formatStr string, v ...interface{}) {
	log.Println(format( "WARN", fmt.Sprintf(formatStr, v...)))
}

func Error(formatStr string, v ...interface{}) {
	log.Println(format("ERROR", fmt.Sprintf(formatStr, v...)))
}

// PrintUser はユーザー向けのメッセージを画面にのみ出力します（ログファイルには記録しません）。
func PrintUser(formatStr string, v ...interface{}) {
	fmt.Printf(formatStr+"\n", v...)
}

// GetLogPath は現在のログファイルのパスを返します。
func GetLogPath(exeBaseDir string) string {
	return filepath.Join(exeBaseDir, "bridge_log.txt")
}
