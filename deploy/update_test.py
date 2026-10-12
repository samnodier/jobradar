"""Exercise release ordering without Docker, network, or production data."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

REVISION = 'a' * 40

class UpdateTest(unittest.TestCase):
    def run_update(self, fail_pull=False):
        task_dir = tempfile.TemporaryDirectory()
        self.addCleanup(task_dir.cleanup)
        root = Path(task_dir.name)
        (root / 'deploy').mkdir()
        (root / 'bin').mkdir()
        shutil.copy(Path(__file__).with_name('update.sh'), root / 'deploy/update.sh')
        (root / '.env').write_text('JOBRADAR_PORT=8090\n')
        scripts = {
            'git': f'''#!/bin/sh
case "$*" in
'branch --show-current') echo main;;
'rev-parse HEAD'|'rev-parse origin/main') echo {REVISION};;
esac
''',
            'docker': '''#!/bin/sh
echo "$*" >> "$ACTION_LOG"
case "$*" in
*'pull --quiet app') [ "$FAIL_PULL" != 1 ];;
*'config --images app') echo ghcr.io/example/app:$JOBRADAR_TAG;;
'image inspect'*) echo wanted-image;;
'inspect'*) echo old-image;;
*'ps -q app'*) echo existing-app;;
esac
''',
            'curl': f'#!/bin/sh\necho \'{{"revision":"{REVISION}"}}\'\n',
            'flock': '#!/bin/sh\nexit 0\n',
        }
        for name, text in scripts.items():
            p = root / 'bin' / name
            p.write_text(text)
            p.chmod(0o755)
        backup = root / 'deploy/backup.sh'
        backup.write_text('#!/bin/sh\necho backup >> "$ACTION_LOG"\n')
        backup.chmod(0o755)
        env = dict(os.environ, PATH=f"{root / 'bin'}:{os.environ['PATH']}", ACTION_LOG=str(root / 'actions'), FAIL_PULL=str(int(fail_pull)))
        result = subprocess.run(['bash', str(root / 'deploy/update.sh')], env=env, capture_output=True, text=True)
        return result, (root / 'actions').read_text(), (root / 'deploy-state/update.json').read_text()

    def test_unpublished_release_never_stops_running_app(self):
        result, actions, state = self.run_update(True)
        self.assertNotEqual(result.returncode, 0)
        self.assertNotIn('stop app', actions)
        self.assertIn('waiting_for_images', state)

    def test_both_pulls_precede_backup_stop_and_migration(self):
        result, actions, state = self.run_update()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertLess(actions.index('pull --quiet migrate'), actions.index('backup'))
        self.assertLess(actions.index('backup'), actions.index('stop app'))
        self.assertLess(actions.index('stop app'), actions.index('run --rm migrate up'))
        self.assertIn('healthy', state)
        self.assertIn(REVISION, state)

if __name__ == '__main__':
    unittest.main()
