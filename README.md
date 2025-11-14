# Video Transcript Player

Tool phát video với transcript được highlight đồng bộ theo thời gian thực, tương tự như các trang học tiếng Nhật.

## Tính năng

- ✅ Upload video và tự động tạo transcript (Embedded → Whisper hoặc ElevenLabs)
- ✅ Tự động extract subtitles từ video nếu có embedded subtitles
- ✅ Nếu không có subtitle, tự động dùng Whisper hoặc ElevenLabs STT
- ✅ Video player với controls đầy đủ
- ✅ Transcript panel hiển thị bên phải
- ✅ Highlight tự động theo thời gian video phát
- ✅ Click vào transcript để jump đến thời điểm tương ứng
- ✅ UI đẹp, responsive

## Cách lấy Transcript

Tool tự động xử lý transcript theo thứ tự:

### 1. Extract từ Video (Ưu tiên)
- Nếu video có **embedded subtitles** (subtitle được nhúng trong file video)
- Tool sẽ tự động extract bằng ffmpeg
- Format: SRT

### 2. Tạo Transcript bằng Whisper / ElevenLabs
- Nếu video không có embedded subtitles
- Bạn có thể chọn provider trong UI:
  1. **Auto / Whisper:** Extract audio → gọi OpenAI Whisper API → nhận transcript với timestamps
  2. **ElevenLabs:** Extract audio → gọi ElevenLabs Speech-to-Text → nếu lỗi sẽ fallback Whisper

## Yêu cầu

- Go 1.21+
- ffmpeg (để extract subtitles và audio từ video)
- OpenAI API Key (Whisper)
- ElevenLabs API Key (nếu muốn dùng ElevenLabs STT/TTS)

### Cài đặt ffmpeg

**macOS:**
```bash
brew install ffmpeg
```

**Linux:**
```bash
sudo apt-get install ffmpeg
```

**Windows:**
Download từ https://ffmpeg.org/download.html

## Cài đặt và chạy

1. Tạo file `.env` trong thư mục project:
```bash
OPENAI_API_KEY=your_openai_api_key_here
ELEVENLABS_API_KEY=your_elevenlabs_api_key_here
```

2. Cài đặt dependencies:
```bash
go mod tidy
```

3. Chạy server:
```bash
go run main.go
```

4. Mở trình duyệt và truy cập:
```
http://localhost:8080
```

## Sử dụng

1. **Upload Video:**
   - Click nút "Chọn Video" để upload video file
   - Chọn provider trong phần “Chọn loại Transcript” (Auto / Whisper / ElevenLabs)
   - Tool sẽ tự động xử lý:
     - Nếu video có embedded subtitles → extract từ video
     - Nếu không có → dùng provider bạn chọn (ElevenLabs hỗ trợ fallback Whisper)

2. **Xem Transcript:**
   - Video sẽ phát ở panel bên trái
   - Transcript sẽ hiển thị ở panel bên phải
   - Transcript tự động highlight khi video phát đến đoạn đó
   - Click vào bất kỳ dòng transcript nào để jump đến thời điểm đó trong video

## Lưu ý về Transcript

- **Video có embedded subtitles:** Tool sẽ tự động extract, không gọi API
- **Video không có subtitles:** Tool dùng provider bạn chọn
- **Fallback:** Nếu chọn ElevenLabs nhưng lỗi, tool tự động fallback sang Whisper
- **API Keys:** Cần `OPENAI_API_KEY` và/hoặc `ELEVENLABS_API_KEY` trong `.env`
- **Chi phí:** Áp dụng theo pricing của từng provider (OpenAI / ElevenLabs)

## Cấu trúc project

```
video-transcript/
├── main.go              # Backend server (Gin)
├── go.mod               # Go dependencies
├── templates/
│   └── index.html       # Frontend UI
├── uploads/             # Thư mục lưu video và subtitles (tự động tạo)
└── README.md
```

## API Endpoints

- `GET /` - Trang chủ
- `POST /api/transcript` - Upload video và tạo transcript (Embedded → Whisper/ElevenLabs)
- `GET /uploads/:filename` - Serve video/subtitle/TTS files

## Lưu ý

- Tool tự động chọn phương pháp dựa vào lựa chọn của bạn (Embedded → Provider)
- Whisper / ElevenLabs API cần internet và API key hợp lệ
- File audio tạm sẽ được tự động xóa sau khi xử lý

