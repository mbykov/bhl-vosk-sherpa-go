#!/usr/bin/env python3
"""
Тестовый инструмент для ASR модуля на Python
Использование:
  python cmd/test.py                    # быстрая проверка файлов
  python cmd/test.py --run              # запустить распознавание
  python cmd/test.py --wav file.wav     # распознать конкретный файл
"""

import json
import argparse
import subprocess
import sys
import os
import tempfile
from pathlib import Path

def load_config(config_path="config.json"):
    """Загружает конфигурацию из JSON"""
    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    except FileNotFoundError:
        print(f"❌ Файл конфигурации {config_path} не найден")
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"❌ Ошибка парсинга JSON: {e}")
        sys.exit(1)

def check_files(config, wav_override=None):
    """Проверяет существование необходимых файлов"""
    model_path = Path(config['model_path'])
    am_onnx = model_path / 'am-onnx'
    lang = model_path / 'lang'

    if not model_path.exists():
        print(f"❌ Модель не найдена: {model_path}")
        return False

    if not (am_onnx / 'encoder.onnx').exists():
        print(f"❌ Файл encoder.onnx не найден в {am_onnx}")
        return False

    if not (lang / 'tokens.txt').exists():
        print(f"❌ Файл tokens.txt не найден в {lang}")
        return False

    wav_file = wav_override or config['test_wav']
    if not Path(wav_file).exists():
        print(f"❌ WAV файл не найден: {wav_file}")
        return False

    print(f"✅ Модель найдена: {model_path}")
    print(f"✅ WAV файл: {wav_file}")
    return True

def run_go_test(config, wav_file):
    """Запускает Go тест"""
    print("\n" + "="*60)
    print("🚀 Запуск ASR теста на Go")
    print("="*60 + "\n")

    # Собираем Go программу
    go_bin = "bin/asr_test"
    os.makedirs("bin", exist_ok=True)

    if not os.path.exists(go_bin):
        print("🔧 Компиляция Go программы...")
        subprocess.run(["go", "build", "-o", go_bin, "cmd/run.go"], check=True)

    # Запускаем тест
    cmd = [go_bin, "-config", "config.json"]
    if wav_file != config['test_wav']:
        cmd.extend(["-wav", wav_file])

    return subprocess.run(cmd).returncode

def quick_test():
    """Быстрый тест без Go (просто проверка файлов)"""
    config = load_config()

    print("📋 Информация о конфигурации:")
    print(f"  Модель: {config['model_path']}")
    print(f"  Тестовый WAV: {config['test_wav']}")
    print(f"  Частота: {config['sample_rate']} Hz")

    model_base = Path(config['model_path'])
    if model_base.exists():
        print(f"\n📁 Содержимое модели:")
        for item in model_base.iterdir():
            if item.is_dir():
                print(f"  📂 {item.name}/")
                files = list(item.iterdir())[:3]
                for sub in files:
                    print(f"     📄 {sub.name}")
            else:
                print(f"  📄 {item.name}")

    wav_path = Path(config['test_wav'])
    if wav_path.exists():
        size = wav_path.stat().st_size / 1024
        print(f"\n✅ Тестовый WAV: {wav_path.name} ({size:.1f} KB)")
    else:
        print(f"\n❌ Тестовый WAV не найден!")

def record_and_test():
    """Запись с микрофона и тестирование"""
    try:
        import pyaudio
        import wave
    except ImportError:
        print("❌ Для записи с микрофона установите: pip install pyaudio")
        return False

    print("\n🎤 Запись с микрофона (5 секунд)...")

    # Параметры записи
    FORMAT = pyaudio.paInt16
    CHANNELS = 1
    RATE = 16000
    CHUNK = 1024
    RECORD_SECONDS = 5

    audio = pyaudio.PyAudio()

    # Открываем поток
    stream = audio.open(format=FORMAT, channels=CHANNELS,
                        rate=RATE, input=True,
                        frames_per_buffer=CHUNK)

    print("🎙️ Говорите...")
    frames = []

    for i in range(0, int(RATE / CHUNK * RECORD_SECONDS)):
        data = stream.read(CHUNK)
        frames.append(data)
        print(f"\r⏺️ Запись: {i*CHUNK/RATE:.1f}/{RECORD_SECONDS} сек", end="")

    print("\n\n✅ Запись завершена")

    # Останавливаем поток
    stream.stop_stream()
    stream.close()
    audio.terminate()

    # Сохраняем во временный файл
    with tempfile.NamedTemporaryFile(suffix='.wav', delete=False) as f:
        temp_wav = f.name

    wf = wave.open(temp_wav, 'wb')
    wf.setnchannels(CHANNELS)
    wf.setsampwidth(audio.get_sample_size(FORMAT))
    wf.setframerate(RATE)
    wf.writeframes(b''.join(frames))
    wf.close()

    print(f"💾 Сохранено в: {temp_wav}")

    # Запускаем распознавание
    config = load_config()
    return run_go_test(config, temp_wav)

def main():
    parser = argparse.ArgumentParser(description="Тестирование ASR модуля")
    parser.add_argument("--wav", help="Путь к WAV файлу")
    parser.add_argument("--config", default="config.json", help="Путь к файлу конфигурации")
    parser.add_argument("--record", action="store_true", help="Записать с микрофона и распознать")
    parser.add_argument("--run", action="store_true", help="Запустить распознавание тестового файла")
    args = parser.parse_args()

    print("🎤 Тестер ASR модуля")
    print("-" * 40)

    # Запись с микрофона
    if args.record:
        return record_and_test()

    # Загружаем конфиг
    config = load_config(args.config)

    # Определяем какой WAV использовать
    wav_file = args.wav or config['test_wav']

    # Проверяем файлы
    if not check_files(config, args.wav):
        sys.exit(1)

    # Если нужен только run или указан wav
    if args.run or args.wav:
        return run_go_test(config, wav_file)
    else:
        quick_test()
        return 0

if __name__ == "__main__":
    sys.exit(main())
