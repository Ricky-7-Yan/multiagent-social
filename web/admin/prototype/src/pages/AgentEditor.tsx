import React, { useEffect, useState } from 'react'

export default function AgentEditor(){
  const [agents, setAgents] = useState<any[]>([])
  const [selected, setSelected] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [persona, setPersona] = useState('')
  const [behavior, setBehavior] = useState('{ \"tone\":\"friendly\" }')

  useEffect(()=> {
    fetch('/api/v1/agents')
      .then(r=> r.json())
      .then(d=> setAgents(d))
      .catch(()=> setAgents([{ id: 'agent-1', name: 'Alice', persona: 'music' }]))
  }, [])

  useEffect(()=> {
    if (!selected) {
      setName('')
      setPersona('')
      setBehavior('{ \"tone\":\"friendly\" }')
      return
    }
    // load agent details
    fetch(`/api/v1/agents/${selected}`)
      .then(r=> {
        if (!r.ok) throw new Error('not found')
        return r.json()
      })
      .then(d=> {
        setName(d.name||'')
        setPersona(d.persona||'')
        setBehavior(JSON.stringify(d.behavior_profile||{}, null, 2))
      })
      .catch(()=> {
        // ignore
      })
  }, [selected])

  const create = async () => {
    const res = await fetch('/api/v1/agents', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, persona }) })
    if (res.ok) {
      const j = await res.json()
      setAgents(a=> [...a, { id: j.id, name, persona }])
      setSelected(j.id)
    } else {
      alert('create failed')
    }
  }

  const update = async () => {
    if (!selected) return
    const res = await fetch(`/api/v1/agents/${selected}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, persona }) })
    if (res.ok) {
      setAgents(a=> a.map(x=> x.id===selected ? { ...x, name, persona } : x))
      alert('Updated')
    } else {
      alert('update failed')
    }
  }

  const remove = async () => {
    if (!selected) return
    if (!confirm('Delete agent?')) return
    const res = await fetch(`/api/v1/agents/${selected}`, { method: 'DELETE' })
    if (res.ok) {
      setAgents(a=> a.filter(x=> x.id!==selected))
      setSelected(null)
    } else {
      alert('delete failed')
    }
  }

  return (
    <div style={{padding:24}}>
      <h1>Agent Editor</h1>
      <div style={{display:'flex',gap:12,marginTop:12}}>
        <div style={{flex:1, background:'#fff', borderRadius:12, padding:12}}>
          <div style={{marginBottom:8}}><strong>{name||'New Agent'}</strong> <span style={{color:'#6B6F82', marginLeft:8}}>Model: gpt-3.5</span></div>
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
            <div style={{marginTop:8}}>
              <label>Existing</label>
              <select style={{width:'100%',padding:8,marginTop:6}} value={selected||''} onChange={e=> setSelected(e.target.value||null)}>
                <option value=''>-- new agent --</option>
                {agents.map(a=> <option key={a.id} value={a.id}>{a.name} ({a.id})</option>)}
              </select>
            </div>
            <div style={{marginTop:8}}>Name<input style={{width:'100%',padding:8,marginTop:6}} value={name} onChange={e=>setName(e.target.value)}/></div>
            <div style={{marginTop:8}}>Persona<textarea style={{width:'100%',padding:8,marginTop:6}} rows={4} value={persona} onChange={e=>setPersona(e.target.value)}/></div>
            <div style={{marginTop:8}}>Behavior JSON<textarea style={{width:'100%',padding:8,marginTop:6}} rows={6} value={behavior} onChange={e=>setBehavior(e.target.value)}/></div>
            <div style={{display:'flex', gap:8, justifyContent:'space-between', marginTop:8}}>
              <div>
                <button onClick={create} style={{background:'#7FB6D9',color:'#fff',padding:'8px 12px',borderRadius:8}}>Create</button>
              </div>
              <div style={{display:'flex', gap:8}}>
                <button onClick={update} style={{background:'#E98FA0',color:'#fff',padding:'8px 12px',borderRadius:8}}>Save</button>
                <button onClick={remove} style={{background:'#E26A6A',color:'#fff',padding:'8px 12px',borderRadius:8}}>Delete</button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  )
}

