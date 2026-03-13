# SodaMusic downloader

一个用 Go 语言开发的汽水音乐下载工具，并自动嵌入元数据和封面图片。

## 功能

- 🎵 **音频下载** - 免费音乐完整下载，会员音乐只可下载前30秒
- 🎨 **元数据嵌入** - 自动嵌入歌曲信息、专辑封面等元数据
- 📋 **剪贴板支持** - 可从剪贴板自动读取 URL

## 💡使用方法

从 [Release 页面](https://github.com/noexcs/SodaMusic-downloader/releases/latest) 下载最新版本的 `sodamusic-downloader.exe`。

### 基本用法

1. #### **复制分享链接后直接双击运行（推荐，自动读取剪贴板中的 URL）**

2. #### 命令行

    ```bash
    # 下载歌词文件
    sodamusic-downloader.exe -lyrics "XXX XXX https://qishui.douyin.com/s/xxx/"
    
    # 指定输出目录
    sodamusic-downloader.exe -output "D:\Music" "XXX XXX https://qishui.douyin.com/s/xxx/"
    
    # 同时使用多个选项
    sodamusic-downloader.exe -lyrics -output "D:\Music" "XXX XXX https://qishui.douyin.com/s/xxx/"
    ```

    #### 选项说明
    
    | 参数 | 说明 | 默认值 |
    |------|------|--------|
    | `-lyrics` | 下载 LRC 格式歌词文件（如果可用） | false |
    | `-output` | 音频文件输出目录 | 当前目录 (.) |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目仅供学习和个人使用。请勿用于商业用途。

---

**提示**：本工具仅用于个人学习和研究。请支持正版音乐，尊重艺术家版权。
