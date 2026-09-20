-- Project-local LazyVim settings: database storage is not Go source.
return {
  {
    "neovim/nvim-lspconfig",
    opts = {
      servers = {
        gopls = {
          settings = {
            gopls = {
              directoryFilters = { "-pgdata", "-**/node_modules", "-frontend/dist" },
            },
          },
        },
      },
    },
  },
}
