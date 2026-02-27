package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strings"

    "vosk-go/asr"  // импортируем наш пакет
)

func main() {
    configPath := flag.String("config", "config.json", "путь к файлу конфигурации")
    wavFile := flag.String("wav", "", "путь к WAV файлу (переопределяет путь из конфига)")
    flag.Parse()

    cfg, err := asr.LoadConfig(*configPath)  // asr.LoadConfig
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

    asrModule, err := asr.New(cfg)  // asr.New
    if err != nil {
        log.Fatalf("❌ Ошибка создания ASR модуля: %v", err)
    }
    defer asrModule.Close()

    err = asrModule.ProcessFile(testFile)
    if err != nil {
        log.Fatalf("❌ Ошибка обработки: %v", err)
    }

    fmt.Println("\n✨ Готово!")
}
