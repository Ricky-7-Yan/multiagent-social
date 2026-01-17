import React, { useEffect, useState } from 'react'

function Sparkline({ values }: { values: number[] }) {
  const max = Math.max(...values, 1)
  const points = values.map((v, i) => `${(i / (values.length - 1)) * 100},${100 - (v / max) * 100}`).join(' ')
  return (
    <svg viewBox="0 0 100 100" style={{ width: 120, height: 40 }}>
      <polyline points={points} fill="none" stroke="#7FB6D9" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

export default function Dashboard() {
  const [agents, setAgents] = useState<any[]>([])
  const [convs, setConvs] = useState<number>(0)
  const [tokens, setTokens] = useState<number>(0)

  useEffect(() => {
    // try to fetch agents from backend; fallback to demo
    fetch('/api/v1/agents')
      .then((r) => r.json())
      .then((d) => setAgents(d))
      .catch(() => setAgents([{ id: 'agent-1', name: 'Alice' }, { id: 'agent-2', name: 'Bob' }]))

    // demo stats
    setConvs(24)
    setTokens(34000)
  }, [])

  return (
    <div style={{ padding: 24 }}>
      <h1>Dashboard</h1>
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <div style={{ background: '#fff', padding: 16, borderRadius: 14, minWidth: 180 }}>
          <div style={{ color: '#6B6F82' }}>Active Agents</div>
          <div style={{ fontSize: 20, fontWeight: 700 }}>{agents.length}</div>
          <div style={{ marginTop: 8 }}><Sparkline values={[2,3,4,5,4,6,7]} /></div>
        </div>
        <div style={{ background: '#fff', padding: 16, borderRadius: 14, minWidth: 180 }}>
          <div style={{ color: '#6B6F82' }}>Conversations</div>
          <div style={{ fontSize: 20, fontWeight: 700 }}>{convs}</div>
          <div style={{ marginTop: 8 }}><Sparkline values={[5,4,6,7,6,8,10]} /></div>
        </div>
        <div style={{ background: '#fff', padding: 16, borderRadius: 14, minWidth: 180 }}>
          <div style={{ color: '#6B6F82' }}>Token Usage</div>
          <div style={{ fontSize: 20, fontWeight: 700 }}>{tokens.toLocaleString()}</div>
          <div style={{ marginTop: 8 }}><Sparkline values={[100,200,150,300,250,400,350]} /></div>
        </div>
      </div>
      <section style={{ marginTop: 18 }}>
        <div style={{ background: 'rgba(255,255,255,0.8)', padding: 16, borderRadius: 12 }}>
          <h3 style={{ marginTop: 0 }}>Recent Activity</h3>
          <div style={{ color: '#6B6F82' }}>Timeline of events and quick insights</div>
          <div style={{ marginTop: 12 }}>
            <div style={{ background: '#fff', padding: 12, borderRadius: 10, marginBottom: 8 }}>Agent Alice executed workflow X — <span style={{ color: '#6B6F82' }}>2m ago</span></div>
            <div style={{ background: '#fff', padding: 12, borderRadius: 10 }}>New conversation started: conv-123 — <span style={{ color: '#6B6F82' }}>10m ago</span></div>
          </div>
        </div>
      </section>
    </div>
  )
}

