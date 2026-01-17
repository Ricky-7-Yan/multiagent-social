import React, { useState } from 'react'

export default function AgentEditor(){
  const [name, setName] = useState('Alice')
  const [persona, setPersona] = useState('music, literature')
  const [behavior, setBehavior] = useState('{ "tone":"friendly" }')

  const save = async () => {
    // mock save: call API if available
    try {
      const res = await fetch('/api/v1/agents', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, persona }) })
      if (res.ok) {
        alert('Saved')
      } else {
        alert('Saved (mock)')
      }
    } catch (e) {
      alert('Saved locally (dev)')
    }
  }

  return (
    <div style={{padding:24}}>
      <h1>Agent Editor</h1>
      <div style={{display:'flex',gap:12,marginTop:12}}>
        <div style={{flex:1, background:'#fff', borderRadius:12, padding:12}}>
          <div style={{marginBottom:8}}><strong>{name}</strong> <span style={{color:'#6B6F82', marginLeft:8}}>Model: gpt-3.5</span></div>
          <div style={{height:320, borderRadius:8, background:'linear-gradient(90deg,#fff,#fbf9f6)', padding:12}}>
            <div style={{display:'flex', gap:8, alignItems:'center'}}>
              <div className="node">Input</div>
              <div className="node">LLM</div>
              <div className="node">Tool</div>
            </div>
            <div style={{marginTop:12, color:'#6B6F82'}}>Drag nodes to compose behavior (mock)</div>
          </div>
        </div>
        <aside style={{width:360}}>
          <div style={{background:'#fff',padding:12,borderRadius:12}}>
            <div style={{fontWeight:700}}>Properties</div>
            <div style={{marginTop:8}}>Name<input style={{width:'100%',padding:8,marginTop:6}} value={name} onChange={e=>setName(e.target.value)}/></div>
            <div style={{marginTop:8}}>Persona<textarea style={{width:'100%',padding:8,marginTop:6}} rows={4} value={persona} onChange={e=>setPersona(e.target.value)}/></div>
            <div style={{marginTop:8}}>Behavior JSON<textarea style={{width:'100%',padding:8,marginTop:6}} rows={6} value={behavior} onChange={e=>setBehavior(e.target.value)}/></div>
            <div style={{textAlign:'right', marginTop:8}}><button style={{background:'#E98FA0',color:'#fff',padding:'8px 12px',borderRadius:8}} onClick={save}>Save</button></div>
          </div>
        </aside>
      </div>
    </div>
  )
}

