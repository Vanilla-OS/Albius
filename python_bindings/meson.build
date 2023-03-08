sudo = find_program('sudo')
python = import('python').find_installation('python3').full_path()

envdata = environment()
envdata.append('ALBIUS_SO_PATH', meson.source_root() + '/albius.so')

test('unit_test',
  sudo,
  args: [ '-E', python, '-m', 'unittest', 'testing/test.py' ],
  depends: albius,
  workdir: meson.current_source_dir(),
  env: envdata,
)