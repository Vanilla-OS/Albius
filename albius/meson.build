subdir('lib')

go_cmd = find_program('go')
go_build_path = meson.current_source_dir() + '/..'

r = run_command(['find', '.', '-name', '\'*.go\''], check: true)
go_files = [
  r.stdout().strip().split(),
  '../albius.go',
  '../go.mod',
]

static_lib_dir = meson.build_root() + '/albius/lib'

albius_env = environment()
albius_env.set('CGO_CFLAGS', '-Ilib')
albius_env.set('CGO_LDFLAGS', '-L' + static_lib_dir + ' -lalbius_internal')

albius = custom_target(
  'albius',
  output: 'albius.so',
  env: albius_env,
  input: go_files,
  depends: internal_lib,
  command: [ go_cmd, 'build', '-o', '@OUTPUT@', '-buildmode=c-shared', go_build_path ],
  install: true,
  install_dir: go_build_path,
)