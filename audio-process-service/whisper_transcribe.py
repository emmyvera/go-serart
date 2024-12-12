# whisper_transcribe.py
import sys
import whisper
import io

# Ensure UTF-8 encoding for stdout and stderr
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8")
sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding="utf-8")

def transcribe(audio_path):
    model = whisper.load_model("base") 
    audio = whisper.load_audio(audio_path)
    audio = whisper.pad_or_trim(audio)
    options = whisper.DecodingOptions(language='english')
    mel = whisper.log_mel_spectrogram(audio, n_mels=model.dims.n_mels).to(model.device)
    result = whisper.decode(model, mel, options)
    text = result.text
    print(text)
    
if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python whisper_transcribe.py <audio_file>")
        sys.exit(1)

    audio_file = sys.argv[1]
    transcribe(audio_file)


