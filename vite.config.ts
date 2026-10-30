import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig(({ command }) => {
  const repositoryName = process.env.GITHUB_REPOSITORY?.split('/').pop() ?? 'IECSECompTeam'

  return {
    base: command === 'serve' ? '/IECSECompTeam/' : `/${repositoryName}/`,
    plugins: [react()],
  }
})
