import WS from 'jest-websocket-mock';
import { v4 as uuidv4 } from 'uuid';

import { ExperimentSpec } from './experiments';
import { ProjectSpec } from './projects';
import { Stream } from './stream';

const onUpsert = vi.fn();
const onDelete = vi.fn();
const isLoaded = vi.fn();
const url = 'ws://localhost:1234';
const genServer = () => new WS(url, { jsonProtocol: true });

const spec1ws = [1];
const spec2ws = [1, 2];
const spec3exp = [1, 2, 3];
const spec1 = new ProjectSpec(spec1ws);
const spec2 = new ProjectSpec(spec2ws);
const spec3 = new ExperimentSpec(spec3exp);

const setup = () => {
  const server = genServer();
  return { client: new Stream(url, onUpsert, onDelete, isLoaded), server };
};

const cleanup = (arr: Array<WS | Stream>) => {
  for (const a of arr) a.close();
};

describe('stream', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });
  it('auto reconnect', async () => {
    const { server, client } = setup();
    let s = await server.connected;
    expect(s.readyState).toEqual(1);

    server.error();
    const newServer = genServer();
    s = await newServer.connected;

    expect(s.readyState).toEqual(1);

    cleanup([client, server, newServer]);
  });
  it('duplicated subscription', async () => {
    const { server, client } = setup();
    await server.connected;

    // Duplicated subscribe won't send twice
    client.subscribe(spec1);
    client.subscribe(spec1);

    await expect(server).toReceiveMessage({
      known: {
        projects: '',
      },
      subscribe: {
        projects: {
          project_ids: [],
          since: 0,
          workspace_ids: spec1ws,
        },
      },
      sync_id: '1',
    });
    server.send({ complete: false, sync_id: '1' });
    server.send({ complete: true, sync_id: '1' });

    client.subscribe(spec2);
    await expect(server).toReceiveMessage({
      known: {
        projects: '',
      },
      subscribe: {
        projects: {
          project_ids: [],
          since: 0,
          workspace_ids: spec2ws,
        },
      },
      sync_id: '2',
    });

    cleanup([client, server]);
  });
  it('recieve message', async () => {
    const { server, client } = setup();
    await server.connected;
    const id = uuidv4();
    client.subscribe(spec1, id);
    server.send({ complete: false, sync_id: '1' });
    server.send({ projects_deleted: '3-5' });
    server.send({ project: { id: 6 } });

    expect(onDelete).toHaveBeenCalledWith('projects', [3, 4, 5]);
    expect(onUpsert).toHaveBeenCalledWith({ project: { id: 6 } });

    server.send({ complete: true, sync_id: '1' });

    expect(isLoaded).toHaveBeenCalledWith([id]);
    cleanup([client, server]);
  });
  it('switch subscription', async () => {
    const { server, client } = setup();
    await server.connected;

    client.subscribe(spec1);
    server.send({ complete: false, sync_id: '1' });
    server.send({ project: { id: 4 } });

    expect(onUpsert).toHaveBeenNthCalledWith(1, { project: { id: 4 } });
    // Create a new subscription, but this subscription won't be sent before the previous sync ends
    client.subscribe(spec2);
    // This message is kept because the initial sync has not completed
    server.send({ project: { id: 5 } });

    expect(onUpsert).toHaveBeenNthCalledWith(2, { project: { id: 5 } });
    server.send({ complete: true, sync_id: '1' });
    // This message is skipped because new subscription has been sent
    server.send({ project: { id: 6 } });
    server.send({ complete: false, sync_id: '2' });
    server.send({ project: { id: 7 } });

    expect(onUpsert).toHaveBeenNthCalledWith(3, { project: { id: 7 } });

    cleanup([client, server]);
  });
  it('connection break', async () => {
    const { server, client } = setup();
    await server.connected;

    client.subscribe(spec1);
    client.subscribe(spec2);

    // server closed right after subsciption sent
    server.close();
    let newServer = genServer();
    await newServer.connected;
    // new server should receive previously sent subscription, with sync_id bumped
    await expect(newServer).toReceiveMessage({
      known: {
        projects: '',
      },
      subscribe: {
        projects: {
          project_ids: [],
          since: 0,
          workspace_ids: spec1ws,
        },
      },
      sync_id: '2',
    });

    newServer.send({ complete: false, sync_id: '2' });
    // server closed after sync started but before sync completed
    newServer.close();
    newServer = genServer();
    await newServer.connected;
    // new server should receive previously interupted subscription, with sync_id bumped
    await expect(newServer).toReceiveMessage({
      known: {
        projects: '',
      },
      subscribe: {
        projects: {
          project_ids: [],
          since: 0,
          workspace_ids: spec1ws,
        },
      },
      sync_id: '3',
    });
    newServer.send({ complete: true, sync_id: '2' });
    // server closed after sync completed
    newServer.close();
    newServer = genServer();
    await newServer.connected;
    // new server should only recieve the second subscription
    await expect(newServer).toReceiveMessage({
      known: {
        projects: '',
      },
      subscribe: {
        projects: {
          project_ids: [],
          since: 0,
          workspace_ids: spec2ws,
        },
      },
      sync_id: '4',
    });

    cleanup([client, newServer]);
  });
  it('multiple entities', async () => {
    const { server, client } = setup();
    await server.connected;

    const proId = uuidv4();
    const expId = uuidv4();
    client.subscribe(spec1, proId);
    client.subscribe(spec3, expId);

    await expect(server).toReceiveMessage({
      known: {
        projects: '',
      },
      subscribe: {
        projects: {
          project_ids: [],
          since: 0,
          workspace_ids: spec1ws,
        },
      },
      sync_id: '1',
    });
    server.send({ complete: false, sync_id: '1' });
    server.send({ project: { id: 1, seq: 1 } });
    expect(onUpsert).toHaveBeenNthCalledWith(1, { project: { id: 1, seq: 1 } });
    server.send({ complete: true, sync_id: '1' });

    expect(isLoaded).toHaveBeenCalledWith([proId]);

    // Second subscription sent only after first subscription loaded
    await expect(server).toReceiveMessage({
      known: {
        experiments: '',
        projects: '1',
      },
      subscribe: {
        experiments: {
          experiment_ids: spec3exp,
          since: 0,
        },
        projects: {
          project_ids: [],
          since: 1,
          workspace_ids: spec1ws,
        },
      },
      sync_id: '2',
    });
    server.send({ complete: false, sync_id: '2' });
    server.send({ complete: true, sync_id: '2' });
    // Both subscriptions loaded
    expect(isLoaded).toHaveBeenCalledWith([proId, expId]);

    // Server sends some updates and we update the keyCache
    server.send({ project: { id: 2, seq: 3 } });
    expect(onUpsert).toHaveBeenNthCalledWith(2, { project: { id: 2, seq: 3 } });
    server.send({ experiment: { id: 4, seq: 5 } });
    server.send({ experiment: { id: 5, seq: 2 } });
    server.send({ experiment: { id: 6, seq: 2 } });

    // Change subscription for projects, since spec2 includes spec1, so keyCache for spec1 is still valid.
    client.subscribe(spec2);

    // Subscription for experiments stay the same, subscription for projects got reset.
    await expect(server).toReceiveMessage({
      known: {
        experiments: '4-6',
        projects: '1-2',
      },
      subscribe: {
        experiments: {
          experiment_ids: spec3exp,
          since: 5,
        },
        projects: {
          project_ids: [],
          since: 3,
          workspace_ids: spec2ws,
        },
      },
      sync_id: '3',
    });
    server.send({ complete: false, sync_id: '3' });
    server.send({ complete: true, sync_id: '3' });
    server.send({ experiments_deleted: '5' });

    client.subscribe(spec1);

    await expect(server).toReceiveMessage({
      known: {
        experiments: '4,6',
        projects: '1-2',
      },
      subscribe: {
        experiments: {
          experiment_ids: spec3exp,
          since: 5,
        },
        projects: {
          project_ids: [],
          since: 3,
          workspace_ids: spec1ws,
        },
      },
      sync_id: '4',
    });
    server.send({ complete: false, sync_id: '4' });
    server.send({ complete: true, sync_id: '4' });

    const newId = uuidv4();
    client.subscribe(spec1, newId);
    expect(isLoaded).toHaveBeenCalledWith([newId]);

    cleanup([client, server]);
  });
});
