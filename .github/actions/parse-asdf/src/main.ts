import * as core from '@actions/core'
import path from 'path'
import {parseToolVersions} from './asdf'

async function run(): Promise<void> {
  try {
    const tools = await parseToolVersions(
      path.join(process.cwd(), '.tool-versions')
    )

    core.setOutput('tools', JSON.stringify(tools))
  } catch (error) {
    core.setFailed(error.message)
  }
}

run()
