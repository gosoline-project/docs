import React from 'react';
import { useState } from 'react';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import LayersIcon from '@mui/icons-material/Layers';
import TerminalIcon from '@mui/icons-material/Terminal';
import Grid from '@mui/material/Grid';
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import ThemeCodeBlock from '@theme/CodeBlock';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import { useColorMode } from '@docusaurus/theme-common'
import useBaseUrl from '@docusaurus/useBaseUrl';

export function CopyToClipboardButton({ text, label = 'Copy to clipboard' }) {
    const [status, setStatus] = useState('idle');

    const copy = async () => {
        try {
            if (navigator.clipboard && window.isSecureContext) {
                await navigator.clipboard.writeText(text);
            } else {
                const textarea = document.createElement('textarea');
                textarea.value = text;
                textarea.style.position = 'fixed';
                textarea.style.left = '-9999px';
                document.body.appendChild(textarea);
                textarea.focus();
                textarea.select();
                document.execCommand('copy');
                document.body.removeChild(textarea);
            }

            setStatus('copied');
            window.setTimeout(() => setStatus('idle'), 2000);
        } catch (err) {
            setStatus('failed');
            window.setTimeout(() => setStatus('idle'), 3000);
        }
    };

    return (
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', margin: '1rem 0' }}>
            <Button variant="contained" onClick={copy}>{label}</Button>
            {status === 'copied' && <span>Copied.</span>}
            {status === 'failed' && <span>Copy failed. Select the instructions manually from the source file.</span>}
        </div>
    );
}

export function PrimaryUseCases() {
    const { colorMode } = useColorMode();
    const isDarkTheme = colorMode === 'dark';

    const darkTheme = createTheme({
      palette: {
        mode: isDarkTheme ? 'dark' : 'light',
      }
    })

    return (
      <ThemeProvider theme={darkTheme}>
        <Grid container spacing={4}>

        <Grid size={{ xs: 12, md: 4 }}>
            <Card style={{ height: '100%' }}>
                <CardHeader title="HTTP Server" avatar={<CloudQueueIcon />} titleTypographyProps={{variant:'h6'}} />
                <CardContent>                    
                    Build REST web services with HTTP handling, caching, OAuth, and much more.
                </CardContent>
                <CardActions>
                    <Button size="small" href={useBaseUrl("/category/http-server")}>Get started</Button>
                </CardActions>
            </Card>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
            <Card style={{ height: '100%' }}>
                <CardHeader title="Message Queues" avatar={<LayersIcon />} titleTypographyProps={{variant:'h6'}} />
                <CardContent>
                    Process asynchronous messages from Kafka, Redis, or any other queuing or streaming system.
                </CardContent>
                <CardActions>
                    <Button size="small" href={useBaseUrl("/quickstart/create-a-consumer")}>Get started</Button>
                </CardActions>
            </Card>
        </Grid>

        <Grid size={{ xs: 12, md: 4 }}>
            <Card style={{ height: '100%' }}>
                <CardHeader title="Kernel Module" avatar={<TerminalIcon />} titleTypographyProps={{variant:'h6'}} />
                <CardContent>                                    
                    Implement a kernel module with which you can do anything, using gosoline's logging, configuration, and other solutions.
                </CardContent>
                <CardActions>
                    <Button size="small" href={useBaseUrl("/quickstart/create-an-application")}>Get started</Button>
                </CardActions>
            </Card>
        </Grid>

        </Grid>
      </ThemeProvider>
    )
  }

export function CodeBlock({ children, snippet, ...props }) {
    let code = children.replace(/\n$/, '');
    var foundMatch = false;

    if (snippet) {
      // Find the snippet
      const snippetPattern = new RegExp(
        `(?:\/\/|#) snippet-start: ${snippet}\s*\n(.*)(?:\/\/|#) snippet-end: ${snippet}\s*$`,
        'sm',
      );
      let match = code.match(snippetPattern)
      if (match) {
        code = match[1];
        foundMatch = true;
      }
    }    

    // Remove all other potential snippet comments as well as the first and last empty lines
    code = code.split('\n').filter(
        function(line) {
            let leftoverCommentsPattern = new RegExp('\w*?(?:\/\/|#) snippet.*\n*?$', 'g')
            return !leftoverCommentsPattern.test(line)
        }
    )
    
    // Drop leading and trailing empty lines
    while (code.length > 0 && code[0].trim() === '') code = code.slice(1);
    while (code.length > 0 && code[code.length - 1].trim() === '') code = code.slice(0, -1);

    code = code.join('\n')

    // Replace leading tabs with 2 spaces
    code = code.replace(/^\t+/gm, match => '  '.repeat(match.length));

    // Strip common leading whitespace from all lines
    const lines = code.split('\n');
    const nonEmptyLines = lines.filter(line => line.trim().length > 0);
    const minIndent = nonEmptyLines.reduce((min, line) => {
        const indent = line.match(/^(\s*)/)[1].length;
        return Math.min(min, indent);
    }, Infinity);
    
    if (minIndent > 0 && minIndent !== Infinity) {
        code = lines.map(line => line.slice(minIndent)).join('\n');
    }

    return <ThemeCodeBlock {...props}>{code}</ThemeCodeBlock>
}
