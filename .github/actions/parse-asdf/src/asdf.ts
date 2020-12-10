import readline from 'readline'
import fs from 'fs'

export async function parseToolVersions(
  file: string
): Promise<Map<string, string>> {
  // const content = fs.readFileSync(file, 'utf8')

  const readInterface = readline.createInterface({
    input: fs.createReadStream(file),
    crlfDelay: Infinity
  })

  const tools = new Map<string, string>()
  for await (const line of readInterface) {
    const tool = line.split(' ')
    tools.set(tool[0], tool[1])
  }

  return tools
}
