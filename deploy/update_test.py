"""Exercise release ordering without Docker, network, or production data."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

REVISION = 'a' * 40

class UpdateTest(unittest.TestCase):
    def run_update(self, fail_pull=False, same_image=False, fail_health=False, fail_git=False, pin=None):
        task_dir = tempfile.TemporaryDirectory()
        self.addCleanup(task_dir.cleanup)
        root = Path(task_dir.name)
        (root / 'deploy').mkdir()
        (root / 'bin').mkdir()
        shutil.copy(Path(__file__).with_name('update.sh'), root / 'deploy/update.sh')
        (root / '.env').write_text('JOBRADAR_PORT=8090\n' + (f'JOBRADAR_TAG={pin}\n' if pin is not None else ''))
        scripts = {
            'git': f'''#!/bin/sh
case "$*" in
'fetch --quiet origin main') [ \"$FAIL_GIT\" != 1 ];;
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
'inspect'*) if [ \"$SAME_IMAGE\" = 1 ]; then echo wanted-image; else echo old-image; fi;;
*'ps -q app'*) echo existing-app;;
esac
''',
            'curl': f'#!/bin/sh\n[ \"$FAIL_HEALTH\" != 1 ] || exit 22\necho \'{{"revision":"{REVISION}"}}\'\n',
            'flock': '#!/bin/sh\nexit 0\n',
        }
        for name, text in scripts.items():
            p = root / 'bin' / name
            p.write_text(text)
            p.chmod(0o755)
        backup = root / 'deploy/backup.sh'
        backup.write_text('#!/bin/sh\necho backup >> "$ACTION_LOG"\n')
        backup.chmod(0o755)
        env = dict(os.environ, PATH=f"{root / 'bin'}:{os.environ['PATH']}", ACTION_LOG=str(root / 'actions'), FAIL_PULL=str(int(fail_pull)), SAME_IMAGE=str(int(same_image)), FAIL_HEALTH=str(int(fail_health)), FAIL_GIT=str(int(fail_git)))
        result = subprocess.run(['bash', str(root / 'deploy/update.sh')], env=env, capture_output=True, text=True)
        return result, ((root / 'actions').read_text() if (root / 'actions').exists() else ''), (root / 'deploy-state/update.json').read_text()

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

    def test_unchanged_unhealthy_image_keeps_failure_visible(self):
        result, actions, state = self.run_update(same_image=True, fail_health=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('failed', state)
        self.assertNotIn('stop app', actions)
        self.assertNotIn('backup', actions)

    def test_unchanged_healthy_image_does_not_redeploy(self):
        result, actions, state = self.run_update(same_image=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('healthy', state)
        self.assertNotIn('stop app', actions)

    def test_invalid_pin_fails_before_pulling_or_stopping(self):
        for pin in ['abc123', 'latest', 'A' * 40]:
            with self.subTest(pin=pin):
                result, actions, state = self.run_update(pin=pin)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('failed', state)
                self.assertNotIn('pull', actions)
                self.assertNotIn('stop app', actions)

    def test_git_fetch_failure_records_failed_state(self):
        result, actions, state = self.run_update(fail_git=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('failed', state)
        self.assertNotIn('pull', actions)

if __name__ == '__main__':
    unittest.main()
