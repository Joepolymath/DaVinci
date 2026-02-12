import { useState, useRef } from 'react'

interface ChatStreamDelta {
  content: string
  done: boolean
  finish_reason?: string
}

export default function App() {
  const [prompt, setPrompt] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [response, setResponse] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const abortRef = useRef<(() => void) | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleSend = async () => {
    if (!prompt.trim() || isStreaming) return

    const content = prompt.trim()
    setPrompt('')
    setResponse('')
    setIsStreaming(true)

    let res: Response
    try {
      res = await fetch('http://localhost:8094/api/chats/stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role: 'user', content }),
      })
    } catch {
      setResponse('Error: Could not reach the server.')
      setIsStreaming(false)
      return
    }

    if (!res.ok || !res.body) {
      setResponse(`Error: ${res.status} ${res.statusText}`)
      setIsStreaming(false)
      return
    }

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let streamDone = false

    abortRef.current = () => reader.cancel()

    try {
      while (!streamDone) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice(6).trim()
          if (data === '[DONE]') {
            streamDone = true
            break
          }
          try {
            const delta: ChatStreamDelta = JSON.parse(data)
            if (delta.content) {
              setResponse(prev => prev + delta.content)
            }
          } catch {
            // ignore parse errors
          }
        }
      }
    } finally {
      setIsStreaming(false)
      abortRef.current = null
    }
  }

  const handleStop = () => {
    abortRef.current?.()
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 flex flex-col items-center px-4 py-12">
      {/* Header */}
      <h1 className="text-4xl font-bold tracking-tight mb-10">
        Da<span className="text-indigo-400">Vinci</span>
      </h1>

      <div className="w-full max-w-3xl flex flex-col gap-4">
        {/* Response area */}
        <div className="min-h-72 max-h-[60vh] bg-gray-900 rounded-xl border border-gray-800 p-5 text-sm leading-relaxed text-gray-200 whitespace-pre-wrap overflow-y-auto">
          {response ? (
            <>
              {response}
              {isStreaming && (
                <span className="inline-block w-2 h-4 bg-indigo-400 ml-0.5 animate-pulse align-middle" />
              )}
            </>
          ) : (
            <span className="text-gray-600">Response will appear here...</span>
          )}
        </div>

        {/* Input card */}
        <div className="flex flex-col gap-3 bg-gray-900 rounded-xl border border-gray-800 p-4">
          <textarea
            className="w-full bg-transparent text-sm text-gray-100 placeholder-gray-600 resize-none outline-none min-h-24"
            placeholder="Type your prompt here... (Enter to send, Shift+Enter for newline)"
            value={prompt}
            onChange={e => setPrompt(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={isStreaming}
          />

          <div className="flex items-center justify-between gap-3 pt-1 border-t border-gray-800">
            {/* File upload */}
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-300 transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="w-4 h-4 shrink-0"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
              </svg>
              <span className="truncate max-w-48">{file ? file.name : 'Attach file'}</span>
            </button>
            <input
              ref={fileInputRef}
              type="file"
              className="hidden"
              onChange={e => setFile(e.target.files?.[0] ?? null)}
            />

            {/* Action buttons */}
            <div className="flex items-center gap-2 shrink-0">
              {file && (
                <button
                  type="button"
                  onClick={() => setFile(null)}
                  className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
                >
                  Remove file
                </button>
              )}
              {isStreaming && (
                <button
                  type="button"
                  onClick={handleStop}
                  className="px-4 py-2 rounded-lg text-sm font-medium bg-gray-700 hover:bg-gray-600 text-gray-200 transition-colors"
                >
                  Stop
                </button>
              )}
              <button
                type="button"
                onClick={handleSend}
                disabled={!prompt.trim() || isStreaming}
                className="px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed text-white transition-colors"
              >
                Send
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
