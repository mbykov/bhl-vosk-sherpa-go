package vosk

import (
    "encoding/json"
    "fmt"
    "os"
    "sync"
    "time"
    "reflect"

    "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

// Config структура для загрузки из JSON
type Config struct {
    ModelPath   string `json:"model_path"`
    TestWav     string `json:"test_wav"`
    SampleRate  int    `json:"sample_rate"`
    FeatureDim  int    `json:"feature_dim"`
    ChunkMs     int    `json:"chunk_ms"`
}

// ASRModule представляет модуль распознавания речи
type ASRModule struct {
    recognizer *sherpa_onnx.OnlineRecognizer
    stream     *sherpa_onnx.OnlineStream
    mu         sync.Mutex
    sampleRate int
    config     Config
    useStub    bool
}

// Result представляет результат распознавания
type Result struct {
    Text        string
    IsFinal     bool
}

// New создает новый экземпляр ASR модуля
func New(cfg Config) (*ASRModule, error) {
    module, err := newReal(cfg)
    if err == nil {
        return module, nil
    }

    fmt.Printf("⚠️  Sherpa-onnx не загружен (%v), использую заглушку\n", err)
    return &ASRModule{
        config:  cfg,
        useStub: true,
    }, nil
}

// newReal создает реальный ASR модуль
func newReal(cfg Config) (*ASRModule, error) {
    // Проверяем существование файлов
    encoderPath := cfg.ModelPath + "/am-onnx/encoder.onnx"
    decoderPath := cfg.ModelPath + "/am-onnx/decoder.onnx"
    joinerPath := cfg.ModelPath + "/am-onnx/joiner.onnx"
    tokensPath := cfg.ModelPath + "/lang/tokens.txt"

    for _, path := range []string{encoderPath, decoderPath, joinerPath, tokensPath} {
        if _, err := os.Stat(path); err != nil {
            return nil, fmt.Errorf("file not found: %s", path)
        }
    }

    // Создаем конфигурацию как в старом проекте
    recognizerConfig := sherpa_onnx.OnlineRecognizerConfig{}

    // Настройка FeatConfig
    recognizerConfig.FeatConfig.SampleRate = cfg.SampleRate
    recognizerConfig.FeatConfig.FeatureDim = cfg.FeatureDim

    // Настройка ModelConfig
    recognizerConfig.ModelConfig.Tokens = tokensPath
    recognizerConfig.ModelConfig.Transducer.Encoder = encoderPath
    recognizerConfig.ModelConfig.Transducer.Decoder = decoderPath
    recognizerConfig.ModelConfig.Transducer.Joiner = joinerPath
    recognizerConfig.ModelConfig.ModelType = "zipformer2"
    recognizerConfig.ModelConfig.Debug = 0
    recognizerConfig.ModelConfig.NumThreads = 1
    recognizerConfig.ModelConfig.Provider = "cpu"

    // Настройка декодирования
    recognizerConfig.DecodingMethod = "greedy_search"
    recognizerConfig.MaxActivePaths = 4
    recognizerConfig.EnableEndpoint = 1

    // Создаем распознаватель
    recognizer := sherpa_onnx.NewOnlineRecognizer(&recognizerConfig)
    if recognizer == nil {
        return nil, fmt.Errorf("failed to create recognizer")
    }

    // Создаем поток
    stream := sherpa_onnx.NewOnlineStream(recognizer)
    if stream == nil {
        sherpa_onnx.DeleteOnlineRecognizer(recognizer)
        return nil, fmt.Errorf("failed to create stream")
    }

    return &ASRModule{
        recognizer: recognizer,
        stream:     stream,
        sampleRate: cfg.SampleRate,
        config:     cfg,
        useStub:    false,
    }, nil
}

// WriteAudio отправляет аудио в распознаватель
func (m *ASRModule) WriteAudio(pcm []byte) error {
    if m.useStub {
        return nil
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    if m.stream == nil {
        return fmt.Errorf("stream is closed")
    }

    // Конвертируем []byte в []float32
    samples := make([]float32, len(pcm)/2)
    for i := 0; i < len(pcm); i += 2 {
        sample := int16(pcm[i]) | int16(pcm[i+1])<<8
        samples[i/2] = float32(sample) / 32768.0
    }

    // Отправляем в поток
    m.stream.AcceptWaveform(m.sampleRate, samples)

    return nil
}

// GetResult получает результат распознавания
func (m *ASRModule) GetResult() (Result, error) {
    if m.useStub {
        return Result{}, nil
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    if m.recognizer == nil || m.stream == nil {
        return Result{}, fmt.Errorf("recognizer or stream is nil")
    }

    // Декодируем поток если готов
    if m.recognizer.IsReady(m.stream) {
        m.recognizer.Decode(m.stream)
    }

    // Получаем результат
    result := m.recognizer.GetResult(m.stream)
    if result == nil {
        return Result{}, nil
    }

    // Получаем текст через рефлексию (работает!)
    text := ""
    v := reflect.ValueOf(result)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    if field := v.FieldByName("Text"); field.IsValid() {
        text = field.String()
    }

    // Проверяем, закончилась ли фраза
    isFinal := m.recognizer.IsEndpoint(m.stream)

    // Если это финальный результат и есть текст, сбрасываем поток
    if isFinal && text != "" {
        m.recognizer.Reset(m.stream)
    }

    return Result{
        Text:    text,
        IsFinal: isFinal,
    }, nil
}

// ProcessFile обрабатывает WAV файл
func (m *ASRModule) ProcessFile(wavPath string) error {
    if m.useStub {
        fmt.Printf("📊 Заглушка: обрабатываю %s\n", wavPath)
        time.Sleep(2 * time.Second)
        fmt.Println("\n🎯 ИТОГ: привет мир (заглушка)")
        return nil
    }

    // Читаем WAV файл
    wavData, err := os.ReadFile(wavPath)
    if err != nil {
        return fmt.Errorf("error reading WAV file: %v", err)
    }

    // Пропускаем WAV заголовок (44 байта для стандартного WAV)
    if len(wavData) < 44 {
        return fmt.Errorf("file too small")
    }
    audioData := wavData[44:]

    fmt.Printf("📊 Распознаю %s (%d байт аудио)...\n\n", wavPath, len(audioData))

    // Рассчитываем размер чанка в байтах (16 бит = 2 байта на сэмпл)
    chunkBytes := m.config.SampleRate * 2 * m.config.ChunkMs / 1000

    // Подаем аудио частями
    for i := 0; i < len(audioData); i += chunkBytes {
        end := i + chunkBytes
        if end > len(audioData) {
            end = len(audioData)
        }

        if err := m.WriteAudio(audioData[i:end]); err != nil {
            return fmt.Errorf("error writing audio: %v", err)
        }

        // Получаем и выводим результат
        result, err := m.GetResult()
        if err != nil {
            continue
        }

        if result.Text != "" {
            if result.IsFinal {
                fmt.Printf("\n✅ ФИНАЛ: %s\n", result.Text)
            } else {
                fmt.Printf("\r🔄 ПРОМЕЖ: %-50s", result.Text)
            }
        }

        // Ждем немного, имитируя реальное время
        time.Sleep(time.Duration(m.config.ChunkMs) * time.Millisecond)
    }

    // Ждем финального результата после окончания аудио
    fmt.Println("\n\n⏳ Ожидание финального результата...")
    for i := 0; i < 20; i++ {
        result, err := m.GetResult()
        if err != nil {
            break
        }
        if result.IsFinal && result.Text != "" {
            fmt.Printf("\n🎯 ИТОГ: %s\n", result.Text)
            return nil
        }
        if result.Text != "" {
            fmt.Printf("\r🔄 ФИНАЛ: %-50s", result.Text)
        }
        time.Sleep(100 * time.Millisecond)
    }

    return nil
}

// Close освобождает ресурсы
func (m *ASRModule) Close() {
    if m.useStub {
        return
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    if m.stream != nil {
        sherpa_onnx.DeleteOnlineStream(m.stream)
        m.stream = nil
    }
    if m.recognizer != nil {
        sherpa_onnx.DeleteOnlineRecognizer(m.recognizer)
        m.recognizer = nil
    }
}

// LoadConfig загружает конфигурацию из JSON файла
func LoadConfig(path string) (Config, error) {
    var cfg Config

    data, err := os.ReadFile(path)
    if err != nil {
        return cfg, fmt.Errorf("error reading config file: %v", err)
    }

    err = json.Unmarshal(data, &cfg)
    if err != nil {
        return cfg, fmt.Errorf("error parsing config JSON: %v", err)
    }

    return cfg, nil
}
