# 📦 Projeto — Setup e Ambiente

Este projeto utiliza:

- Go 1.22+
- Templ (templates)
- Tesseract OCR (via gosseract)
- PDF → Imagem → OCR
- CGO obrigatório

⚠️ O processo de setup é diferente no Linux e no Windows, principalmente por causa do OCR.


⚠️ A pasta `internal/template` contém arquivos `.templ`.  
Eles precisam ser gerados antes de rodar o projeto.

## 1️⃣ Instalar dependências do sistema

# 🐧 Setup no Linux (recomendado)
```bash
sudo apt update
sudo apt install -y \
  tesseract-ocr \
  libtesseract-dev \
  poppler-utils \
  build-essential \
  tesseract-ocr-por \
  tesseract-ocr-spa \
  tesseract-ocr-eng
```

# Windows 
## Download: https://www.msys2.org/

## MSYS2 UCRT64
```bash
  pacman -Syu
  pacman -S mingw-w64-x86_64-gcc
```

## MSYS2 MINGW
```bash
  pacman -S mingw-w64-x86_64-tesseract-ocr mingw-w64-x86_64-leptonica  
```

## windows + R: sysdm.cpl
## add Path: C:\msys64\mingw64\bin

# Download: https://github.com/UB-Mannheim/tesseract/wiki

## windows + R: sysdm.cpl
## add Path: C:\Program Files\Tesseract-OCR

# Download: https://github.com/oschwartz10612/poppler-windows/releases/
## extraia o arquivo em C:// 
## windows + R: sysdm.cpl
## add Path: C:\poppler-25.12.0\Library\bin


# go env CGO_ENABLED
## Se vier 0: setx CGO_ENABLED 1

# 2️⃣ Instalar ferramentas Go
```bash
  go install github.com/go-delve/delve/cmd/dlv@latest
```
```bash
  go install github.com/a-h/templ/cmd/templ@latest
```
# 3️⃣ Gerar os arquivos do Templ
```bash
templ generate 
```

# 4️ Atualizar dependências
```bash
go mod tidy
```

## LGPD e segurança operacional

Use `.env.example` como referência e configure em produção:

- `DOMAIN_URL=https://...`
- `COOKIE_SECURE=true`
- `LOGIN_FINGERPRINT_SECRET` com valor forte e estável
- `PRIVACY_CONTACT_EMAIL` e/ou `DPO_EMAIL`

Rotas úteis para direitos do titular:

- `GET /api/me/export`
- `DELETE /api/me`
- `POST /api/logout`
- `GET /:lang/privacy`

A revisão técnica está em `docs/lgpd-audit.md`.
