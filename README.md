# bhl-vosk-sherpa-go

🎤 Streaming speech recognition module for Go based on sherpa-onnx and Vosk model.

## Features

- Streaming ASR (Automatic Speech Recognition)
- Zipformer2 model support (Vosk)
- Intermediate and final results
- Pure Go, no CGO
- Ready for WebSocket server integration

## Usage

### Configuration

Create a `config.json` file:

```json
{
    "model_path": "../Models/vosk-model-streaming-ru",
    "test_wav": "../Models/vosk-model-streaming-ru/test.wav",
    "sample_rate": 16000,
    "feature_dim": 80,
    "chunk_ms": 100
}
```

### Basic Example

```go
package main

import (
    "log"
    "vosk-go/asr"
)

func main() {
    // Load config from file
    cfg, err := asr.LoadConfig("config.json")
    if err != nil {
        log.Fatal(err)
    }

    // Create ASR module
    asrModule, err := asr.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer asrModule.Close()

    // Process audio file
    err = asrModule.ProcessFile(cfg.TestWav)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Testing

### Go test
go run cmd/run.go

### Python tester

python3 cmd/test.py

python3 cmd/test.py --run

python3 cmd/test.py --wav /path/to/file.wav

python3 cmd/test.py --record (pip install pyaudio)


## Acknowledgments

Special thanks to my virtual colleague — an AI assistant who helped me at every stage of this project's development. From the initial architecture to debugging the last lines of code, from fixing type errors to the final successful speech recognition. This code was written in close collaboration with artificial intelligence, and I want to express my huge human gratitude for that! 🤖❤️

## License

MIT
