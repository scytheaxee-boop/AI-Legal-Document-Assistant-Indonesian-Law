import { useState } from 'react'
import axios from 'axios'
import ReactMarkdown from 'react-markdown'
import './App.css'

interface ChatResponse {
  pertanyaan: string;
  jawaban: string;
  referensi: string;
}

function App() {
  const [pertanyaan, setPertanyaan] = useState('')
  const [loading, setLoading] = useState(false)
  const [hasil, setHasil] = useState<ChatResponse | null>(null)
  const [errorText, setErrorText] = useState('')

  const tanyaHukum = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!pertanyaan.trim()) return

    setLoading(true)
    setErrorText('')
    setHasil(null)

    try {
      const response = await axios.post('/api/tanya', { pertanyaan })
      setHasil(response.data)
    } catch (err: any) {
      console.error(err)
      // Menangkap error 429 (kuota penuh) secara khusus agar terlihat di UI
      if (err.response && err.response.status === 500) {
          setErrorText('Server sibuk atau kuota harian API penuh. Harap tunggu sebentar lalu coba lagi.')
      } else {
          setErrorText('Terjadi kesalahan saat menghubungi asisten hukum.')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="app-container" style={{ maxWidth: '800px', margin: '0 auto', padding: '2rem', fontFamily: 'system-ui' }}>
      <h1 style={{ textAlign: 'center', color: '#2c3e50' }}>⚖️ Legal Document Assistant</h1>
      <p style={{ textAlign: 'center', color: '#7f8c8d' }}>Tanyakan masalah hukum Anda, AI akan menjawab berdasarkan dokumen Undang-Undang.</p>

      <form onSubmit={tanyaHukum} style={{ display: 'flex', gap: '10px', marginTop: '2rem' }}>
        <input 
          type="text" 
          value={pertanyaan}
          onChange={(e) => setPertanyaan(e.target.value)}
          placeholder="Contoh: Apa hukumannya jika mencemarkan nama baik di internet?" 
          style={{ flex: 1, padding: '12px', borderRadius: '8px', border: '1px solid #ccc', fontSize: '16px' }}
          disabled={loading}
        />
        <button 
          type="submit" 
          disabled={loading || !pertanyaan}
          style={{ padding: '12px 24px', borderRadius: '8px', backgroundColor: '#2980b9', color: 'white', border: 'none', cursor: 'pointer', fontWeight: 'bold' }}
        >
          {loading ? 'Menganalisis...' : 'Tanya AI'}
        </button>
      </form>

      {errorText && (
        <div style={{ marginTop: '1rem', padding: '1rem', backgroundColor: '#ffcccc', borderRadius: '8px', color: '#cc0000' }}>
          {errorText}
        </div>
      )}

      {hasil && !loading && (
        <div style={{ marginTop: '2rem', padding: '1.5rem', backgroundColor: '#f9f9f9', borderRadius: '8px', border: '1px solid #ddd' }}>
          <h3 style={{ marginTop: 0, color: '#27ae60' }}>💡 Jawaban Asisten:</h3>
          <div style={{ lineHeight: '1.6', fontSize: '15px' }}>
            <ReactMarkdown>{hasil.jawaban}</ReactMarkdown>
          </div>
          
          <hr style={{ margin: '1.5rem 0', border: '0.5px solid #eee' }} />
          
          <h4 style={{ color: '#7f8c8d', marginBottom: '8px' }}>📚 Referensi Pasal Terkait:</h4>
          <pre style={{ whiteSpace: 'pre-wrap', backgroundColor: '#fff', padding: '1rem', border: '1px solid #eee', borderRadius: '4px', fontSize: '13px', color: '#555' }}>
            {hasil.referensi || "Tidak ada referensi dokumen yang ditemukan."}
          </pre>
        </div>
      )}
    </div>
  )
}

export default App