# veveojapanesechecker

On the app's prompt enter the folder name for GGuide XML files
the app will:

  Process each XML file:
  
    - extract the list of titles
    
      - for each title in the list
      
        - run Veveo search request with the title as a search tearm
        
          - for search result:
          
            - print number of responses
            
            - print each result title with the original search term
            
