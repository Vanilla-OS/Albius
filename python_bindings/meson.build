sudo = find_program('sudo')
python = find_program('python').full_path()

test('test_disk_ops',
  sudo,
  args: [ python, meson.current_source_dir() + '/main.py' ],
  depends: albius,
  workdir: meson.source_root(),
)