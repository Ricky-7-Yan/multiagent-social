import React, { useEffect, useState, useRef } from 'react'

export default function Conversation(){
  const [convId, setConvId] = useState<string>('conv-12345')
  const [messages, setMessages] = useState<any[]>([])
  const [text, setText] = useState('')
  const wsRef = useRef<WebSocket|null>(null)

  useEffect(()=> {
    // load messages
    fetch(`/api/v1/conversations/${encodeURIComponent(convId)}/messages`)
      .then(r=> r.json())
      .then(d=> setMessages(d))
      .catch(()=> setMessages(['Alice: sample message','Bob: sample reply']))

    // try to open websocket
    // prefer EventSource (SSE) for devserver events, fallback to WebSocket then mock
    try {
      const es = new EventSource(`/events/conversations/${encodeURIComponent(convId)}`)
      es.onmessage = (ev) => {
        try { const data = JSON.parse(ev.data); setMessages(m=> [...m, (data.sender ? data.sender+': '+data.content : ev.data)]) } catch(e){ setMessages(m=> [...m, ev.data])}
      }
      es.onerror = () => { es.close() }
      return ()=> { es.close() }
    } catch (e) {
      try {
        const ws = new WebSocket(`ws://${location.host}/ws/conversations/${encodeURIComponent(convId)}`)
        ws.onmessage = (ev)=> {
          try { const data = JSON.parse(ev.data); setMessages(m=> [...m, data.content||ev.data]) } catch(e){ setMessages(m=> [...m, ev.data])}
        }
        ws.onopen = ()=> console.log('ws open')
        wsRef.current = ws
        return ()=> { ws.close() }
      } catch (e2) {
        // fallback: mock real-time by interval
        const id = setInterval(()=> setMessages(m => [...m, 'Mock agent ping: ' + new Date().toLocaleTimeString()]), 5000)
        return ()=> clearInterval(id)
      }
    }
  }, [convId])

  const post = async ()=> {
    if (!convId || !text) return
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

  return (
    <div style={{padding:24}}>
      <h1>Conversation</h1>
      <div style={{display:'flex',gap:12,marginTop:12}}>
        <div style={{flex:1, background:'#fff', borderRadius:12, padding:12}}>
          <div style={{marginBottom:8}}>
            <strong>{convId}</strong>
            <span style={{color:'#6B6F82', marginLeft:8}}>Participants: Alice,Bob</span>
          </div>
          <div style={{minHeight:300}}>
            {messages.map((m, idx) => (
              <div key={idx} className="message fade-in" style={{background: idx%2? '#fff':'#f8f7f3', marginBottom:8}}>
                {typeof m === 'string' ? m : JSON.stringify(m)}
                <div className="meta"><span className="timestamp">now</span></div>
              </div>
            ))}
          </div>
          <div style={{marginTop:8}}>
            <textarea rows={3} style={{width:'100%',padding:8}} placeholder="Type a message..." value={text} onChange={e=>setText(e.target.value)} />
            <div style={{textAlign:'right', marginTop:6}}><button className="btn" onClick={post}>Send</button></div>
            <div style={{marginTop:8, display:'flex', gap:8, alignItems:'center'}}>
              <label style={{fontSize:12, color:'#6B6F82'}}>Real-time</label>
              <input type="checkbox" checked={!!wsRef.current} readOnly />
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

