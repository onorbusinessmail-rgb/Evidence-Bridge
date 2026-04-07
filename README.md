# 🚀 Evidence-Bridge

**PCからAndroidへファイルを爆速で転送する、Go言語製のエンジニア向けツール。**

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
![Platform: Windows](https://img.shields.io/badge/Platform-Windows-0078D6.svg)
![Language: Go](https://img.shields.io/badge/Language-Go-00ADD8.svg)

---

## 💎 特徴
* ⚡ **爆速転送:** ADBを活用し、大容量ファイルもストレスなく転送。
* 🛠️ **簡単操作:** 転送プロセスを自動化し、データ連携をスムーズに。
* 💻 **対応環境:** Windows PC ➔ Android 端末

## 📖 使い方
> [!TIP]
> 初回起動時に設定ファイル (`config.ini`) が自動生成されます。

1. **準備:** Android端末の「USBデバッグ」を有効にし、PCと接続します。
2. **実行:** `main.go` を実行、またはビルドした `exe` を起動します。
3. **設定:** 初回のみ、転送先（メインストレージ / SDカード）を選択。
4. **転送:** `Send` フォルダにファイルを置くと、自動的に転送が開始されます。
5. **確認:** 完了後、`bridge_log.txt` でログを確認できます。

## ⚠️ 制限事項
| 項目 | 内容 |
| :--- | :--- |
| **SDカード判定** | 未装着時に「SDカード」を選択した場合の動作は未考慮です。 |
| **転送方向** | 現状は **PC ➔ Android** の一方通行のみ対応しています。 |

## 🗺️ 今後のロードマップ
- [ ] 🔄 **接続リトライ:** USBの接触不良時、自動で再接続・再開。
- [ ] 🕒 **フォルダ監視:** `Send` フォルダを監視し、指定時刻に自動転送。
- [ ] 🎨 **GUI実装:** ドラッグ&ドロップで操作可能な専用ウィンドウ。

## 🤝 開発への参加
バグ報告や機能拡張のプルリクエストを大歓迎します！
自由にフォークして、あなたのアイディアを形にしてください。

## 🛒 有料版（製品版）について
より高度なサポートと機能を備えたパッケージ版を販売予定です。準備ができ次第、こちらにリンクを掲載します。

---

# 🚀 Evidence-Bridge (English)

A blazing fast tool developed in Go for transferring files from PC to Android.

## ✨ Features
* ⚡ **Blazing Fast:** High-speed file transfer leveraging ADB.
* 🛠️ **Seamless:** Designed for a streamlined data transfer workflow.
* 💻 **Environment:** Supports Windows PC to Android device transfers.

## 📖 Usage
1. **Prepare:** Enable "USB Debugging" on Android and connect to PC.
2. **Run:** Execute `main.go` or launch the compiled `.exe`.
3. **Setup:** On the first run, choose destination (Internal Storage / SD Card).
4. **Transfer:** Drop files into the `Send` folder (works before or during execution).
5. **Verify:** Check `bridge_log.txt` for the results.

## 🗺️ Roadmap
- [ ] **Connection Retry:** Auto-reconnect and resume if USB is interrupted.
- [ ] **Hot Folder:** Scheduled automatic transfers from the `Send` folder.
- [ ] **GUI:** Intuitive drag-and-drop graphical interface.

## 🛒 Commercial Edition
A packaged commercial version is planned for release. Links will be provided here as soon as it's ready.
