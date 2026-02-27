package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strings"

    "vosk-go/vosk"
)

func main() {
    configPath := flag.String("config", "config.json", "путь к файлу конфигурации")
    wavFile := flag.String("wav", "", "путь к WAV файлу")
    flag.Parse()

    cfg, err := vosk.LoadConfig(*configPath)  // было asr.LoadConfig
    if err != nil {
        log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
    }

    testFile := cfg.TestWav
    if *wavFile != "" {
        testFile = *wavFile
    }

    if _, err := os.Stat(testFile); os.IsNotExist(err) {
        log.Fatalf("❌ Файл не найден: %s", testFile)
    }

    fmt.Printf("🔧 Конфигурация:\n")
    fmt.Printf("  Модель: %s\n", cfg.ModelPath)
    fmt.Printf("  Тестовый файл: %s\n", testFile)
    fmt.Printf("  Частота: %d Hz\n", cfg.SampleRate)
    fmt.Printf("  Размер чанка: %d мс\n", cfg.ChunkMs)
    fmt.Println(strings.Repeat("─", 50))

    voskModule, err := vosk.New(cfg)  // было asr.New
    if err != nil {
        log.Fatalf("❌ Ошибка создания Vosk модуля: %v", err)
    }
    defer voskModule.Close()

    err = voskModule.ProcessFile(testFile)
    if err != nil {
        log.Fatalf("❌ Ошибка обработки: %v", err)
    }

    fmt.Println("\n✨ Готово!")
}
