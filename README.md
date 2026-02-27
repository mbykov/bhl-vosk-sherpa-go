# bhl-vosk-sherpa-go

🎤 Streaming speech recognition module for Go based on sherpa-onnx and Vosk model.

## Features

- Streaming ASR (Automatic Speech Recognition)
- Zipformer2 model support (Vosk)
- Intermediate and final results
- Pure Go, no CGO
- Ready for WebSocket server integration

## Usage

```go
cfg := asr.Config{
    ModelPath:  "../Models/vosk-model-streaming-ru",
    SampleRate: 16000,
    FeatureDim: 80,
    ChunkMs:    100,
}

asrModule, _ := asr.New(cfg)
defer asrModule.Close()

// Feed audio chunks
asrModule.WriteAudio(pcmData)

// Get results
result, _ := asrModule.GetResult()
fmt.Printf("Text: %s (final: %v)\n", result.Text, result.IsFinal)
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
