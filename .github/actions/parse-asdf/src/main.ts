import * as core from '@actions/core'
import path from 'path'
import {parseToolVersions} from './asdf'

async function run(): Promise<void> {
  try {
    const file = path.join(process.cwd(), '.tool-versions')
    core.debug(file)
    const tools = await parseToolVersions(file)

    core.warning('All found versions are exported to env variables')
    core.startGroup('.tool-versions')
    for (const [key, value] of tools) {
      core.info(`Gathered '${key}' version ${value}`)
      core.exportVariable(`${key.toUpperCase()}_VERSION`, value)
    }
    core.endGroup()

    core.setOutput('tools', JSON.stringify(tools))
  } catch (error) {
    core.setFailed(error.message)
  }
}

run()
