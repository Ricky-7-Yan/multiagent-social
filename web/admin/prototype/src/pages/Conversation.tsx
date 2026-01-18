import React, { useEffect, useState, useRef } from 'react'

export default function Conversation(){
  const [convId, setConvId] = useState<string>('conv-12345')
  const [messages, setMessages] = useState<any[]>([])
  const [text, setText] = useState('')
  const wsRef = useRef<WebSocket|null>(null)
  const esRef = useRef<EventSource|null>(null)
  const [connected, setConnected] = useState(false)
  const [typing, setTyping] = useState(false)
  const endRef = useRef<HTMLDivElement|null>(null)

  useEffect(()=> {
    // load messages
    fetch(`/api/v1/conversations/${encodeURIComponent(convId)}/messages`)
      .then(r=> r.json())
      .then(d=> setMessages(d))
      .catch(()=> setMessages(['Alice: sample message','Bob: sample reply']))

    // prefer SSE with reconnect/backoff, fallback to WebSocket then mock
    let cancelled = false
    let backoff = 500
    function connectSSE() {
      if (cancelled) return
      try {
        const es = new EventSource(`/events/conversations/${encodeURIComponent(convId)}`)
        esRef.current = es
        es.onopen = () => {
          setConnected(true)
          backoff = 500
        }
        es.onmessage = (ev) => {
          try {
            const data = JSON.parse(ev.data)
            setMessages(m=> [...m, (data.sender ? data.sender+': '+data.content : ev.data)])
          } catch(e) {
            setMessages(m=> [...m, ev.data])
          }
        }
        es.onerror = () => {
          setConnected(false)
          try { es.close() } catch {}
          esRef.current = null
          // reconnect with backoff
          setTimeout(()=> { backoff = Math.min(5000, backoff * 2); connectSSE() }, backoff)
        }
      } catch (err) {
        // fallback to ws
        try {
          const ws = new WebSocket(`ws://${location.host}/ws/conversations/${encodeURIComponent(convId)}`)
          ws.onmessage = (ev)=> {
            try { const data = JSON.parse(ev.data); setMessages(m=> [...m, data.content||ev.data]) } catch(e){ setMessages(m=> [...m, ev.data])}
          }
          ws.onopen = ()=> setConnected(true)
          ws.onclose = ()=> setConnected(false)
          wsRef.current = ws
        } catch (e2) {
          const id = setInterval(()=> setMessages(m => [...m, 'Mock agent ping: ' + new Date().toLocaleTimeString()]), 5000)
          return ()=> clearInterval(id)
        }
      }
    }
    connectSSE()
    return ()=> {
      cancelled = true
      if (esRef.current) try { esRef.current.close() } catch {}
      if (wsRef.current) try { wsRef.current.close() } catch {}
    }
  }, [convId])

  const post = async ()=> {
    if (!convId || !text) return
    setTyping(false)
    try {
      await fetch(`/api/v1/conversations/${encodeURIComponent(convId)}/messages`, { 
        method: 'POST', 
        headers: { 'Content-Type': 'text/plain; charset=utf-8' },
        body: text 
      })
      setMessages(m=> [...m, 'You: '+text])
      setText('')
    } catch (e){ console.error(e) }
  }

  // autoscroll
  useEffect(()=> {
    if (endRef.current) {
      endRef.current.scrollIntoView({ behavior: 'smooth', block: 'end' })
    }
  }, [messages])

  return (
    <div style={{padding:24}}>
      <h1>Conversation</h1>
      <div style={{display:'flex',gap:12,marginTop:12}}>
        <div style={{flex:1, background:'#fff', borderRadius:12, padding:12}}>
          <div style={{marginBottom:8}}>
            <strong>{convId}</strong>
            <span style={{color:'#6B6F82', marginLeft:8}}>Participants: Alice,Bob</span>
          </div>
          <div style={{minHeight:300, maxHeight:480, overflow:'auto', padding:8}}>
            {messages.map((m, idx) => {
              const text = typeof m === 'string' ? m : JSON.stringify(m)
              const isYou = String(text).startsWith('You:')
              return (
                <div key={idx} className={`message ${isYou? 'you' : 'agent'} fade-in`} style={{ marginBottom:8}}>
                  <div className="bubble">{text}</div>
                  <div className="meta"><span className="timestamp">now</span></div>
                </div>
              )
            })}
            <div ref={el => endRef.current = el}></div>
          </div>
          <div style={{marginTop:8}}>
            <textarea rows={3} style={{width:'100%',padding:8}} placeholder="Type a message..." value={text} onChange={e=>setText(e.target.value)} />
            <div style={{textAlign:'right', marginTop:6}}><button className="btn" onClick={post}>Send</button></div>
            <div style={{marginTop:8, display:'flex', gap:8, alignItems:'center'}}>
              <label style={{fontSize:12, color:'#6B6F82'}}>Real-time</label>
              <input type="checkbox" checked={connected} readOnly />
              <div style={{flex:1}} />
              {typing && <div style={{fontSize:12, color:'#6B6F82'}}>typing...</div>}
            </div>
          </div>
        </div>
        <aside style={{width:320}}>
          <div style={{background:'#fff',padding:12,borderRadius:12}}>Participants<ul><li>Alice</li><li>Bob</li></ul></div>
        </aside>
      </div>
    </div>
  )
}

