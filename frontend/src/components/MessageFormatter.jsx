import React from 'react';

const MessageFormatter = ({ content, isStreaming }) => {
  // Detect if content contains code patterns
  const detectCode = (text) => {
    const codePatterns = [
      /class\s+\w+.*:/,           // class definitions
      /def\s+\w+\(/,              // function definitions  
      /import\s+\w+/,             // import statements
      /from\s+\w+\s+import/,      // from imports
      /\w+\s*=\s*\w+\(/,          // variable assignments with function calls
      /if\s+\w+.*:/,              // if statements
      /for\s+\w+\s+in.*:/,        // for loops
      /print\s*\(/,               // print statements
      /return\s+/,                // return statements
      /^\s*#\s/m,                 // comments
      /\w+\.\w+\(/,               // method calls
    ];
    
    return codePatterns.some(pattern => pattern.test(text)) && 
           text.length > 30 && 
           text.includes('\n');
  };

  // Format code with basic syntax highlighting using CSS classes
  const formatCode = (code) => {
    return code
      .replace(/(class\s+\w+)/g, '<span class="keyword">$1</span>')
      .replace(/(def\s+\w+)/g, '<span class="function">$1</span>')
      .replace(/(import|from|return|if|for|in|print)/g, '<span class="keyword">$1</span>')
      .replace(/(#.*$)/gm, '<span class="comment">$1</span>')
      .replace(/('.*?'|".*?")/g, '<span class="string">$1</span>')
      .replace(/\b(\d+)\b/g, '<span class="number">$1</span>');
  };

  // Add line numbers to code
  const addLineNumbers = (code) => {
    const lines = code.split('\n');
    return lines.map((line, index) => (
      `<div class="code-line">
        <span class="line-number">${(index + 1).toString().padStart(2, ' ')}</span>
        <span class="line-content">${line}</span>
      </div>`
    )).join('');
  };

  // Parse content for code blocks and regular text
  const parseContent = (text) => {
    // Split by potential code blocks (look for multi-line content with code patterns)
    const parts = [];
    const paragraphs = text.split('\n\n');
    
    paragraphs.forEach((paragraph, index) => {
      if (detectCode(paragraph)) {
        // This is a code block
        const formattedCode = formatCode(paragraph);
        const codeWithLines = addLineNumbers(formattedCode);
        
        parts.push(
          <div key={`code-${index}`} className="code-block-container">
            <div className="code-block-header">
              <span className="code-language">python</span>
              <button 
                className="copy-button"
                onClick={() => {
                  navigator.clipboard.writeText(paragraph);
                }}
                title="Copy code"
              >
                📋 Copy
              </button>
            </div>
            <div 
              className="code-block"
              dangerouslySetInnerHTML={{ __html: codeWithLines }}
            />
          </div>
        );
      } else {
        // Regular text paragraph
        parts.push(
          <p key={`text-${index}`} className="text-paragraph">
            {paragraph}
          </p>
        );
      }
    });
    
    return parts;
  };

  return (
    <div className={`message-formatter ${isStreaming ? 'streaming' : ''}`}>
      {parseContent(content)}
    </div>
  );
};

export default MessageFormatter;
