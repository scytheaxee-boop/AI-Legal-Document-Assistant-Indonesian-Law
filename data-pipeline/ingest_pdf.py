import os
import re
import time  # <-- Menambahkan modul waktu untuk memberi jeda
from pypdf import PdfReader
import google.generativeai as genai
from supabase import create_client, Client

# --- KONFIGURASI ---
# Ganti dengan API Key masing-masing
GENAI_API_KEY = "AIzaSyABFBdKbwUfNCFF9rNcilzzo7CiV998onI"
SUPABASE_URL = "https://gdupesltihkwiygmfdfe.supabase.co"
SUPABASE_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImdkdXBlc2x0aWhrd2l5Z21mZGZlIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzY1OTA2MjYsImV4cCI6MjA5MjE2NjYyNn0.OmFowyhUa3oagbeVppQN1HG61eqwnGpRX9KJPs3JG8I"

genai.configure(api_key=GENAI_API_KEY)
supabase: Client = create_client(SUPABASE_URL, SUPABASE_KEY)

def extract_text_from_pdf(pdf_path):
    reader = PdfReader(pdf_path)
    text = ""
    for page in reader.pages:
        extracted = page.extract_text()
        if extracted:
            text += extracted.replace('\x00', '') + "\n"
    return text

def chunk_by_pasal(text):
    chunks = re.split(r'(?i)(?=pasal\s+\d+)', text)
    return [c.strip() for c in chunks if len(c.strip()) > 10]

def generate_embedding(text):
    result = genai.embed_content(
        model="models/gemini-embedding-001",
        content=text,
        task_type="retrieval_document",
        output_dimensionality=768  # Kunci agar ukuran vektor cocok dengan database
    )
    return result['embedding']

def main():
    pdf_files = [f for f in os.listdir('.') if f.endswith('.pdf')]
    
    if not pdf_files:
        print("Tidak ada file PDF yang ditemukan di folder ini.")
        return

    print(f"Ditemukan {len(pdf_files)} file PDF. Memulai proses massal...\n")
    
    for pdf_filename in pdf_files:
        print("==================================================")
        print(f"MEMBACA FILE: {pdf_filename}")
        print("==================================================")
        
        raw_text = extract_text_from_pdf(pdf_filename)
        pasal_list = chunk_by_pasal(raw_text)
        
        print(f"-> Ditemukan {len(pasal_list)} potongan dari file ini.\n")
        
        for i, isi in enumerate(pasal_list):
            judul_pasal = isi.split('\n')[0][:50].strip()
            
            print(f"[{i+1}/{len(pasal_list)}] Menyimpan {judul_pasal}...")
            
            try:
                vector = generate_embedding(isi)
                supabase.table("dokumen_hukum").insert({
                    "pasal": f"[{pdf_filename}] {judul_pasal}", 
                    "isi_teks": isi,
                    "embedding": vector
                }).execute()
                
                # --- REM API ---
                # Istirahat 3 detik agar tidak terkena tilang Error 429 dari Google
                time.sleep(3)
                
            except Exception as e:
                print(f"Gagal menyimpan pada bagian {judul_pasal}: {e}")
                
    print("\n✅ SEMUA FILE PDF BERHASIL DIPROSES DAN DISIMPAN KE DATABASE!")

if __name__ == "__main__":
    main()

     ###python ingest_pdf.py