import { useEffect, useState } from 'react'
import './App.css'

interface ToDoSingleContent {
  id: number;
  title: string;
};

interface ToDoListItem {
  key: number;
  content: ToDoSingleContent;
};

type ToDoTable = {
  [index: number]: ToDoListItem;
};

// const todotable: ToDoTable = {
//   0: { key: 1, content: { id: 0, title: '部屋を片付ける' }},
//   1: { key: 2, content: { id: 1, title: '友達にメッセージを返す' }},
//   2: { key: 3, content: { id: 2, title: '昼寝する' }},
//   3: { key: 4, content: { id: 3, title: '夕飯を作る' }},
//   4: { key: 5, content: { id: 4, title: 'ゴールデンレトリーバーと遊ぶ' }},
// };

const ToDoList: React.FC = () => {
  const [table, setTable] = useState<ToDoTable>({});
  const [isInitialized, setIsInitialized] = useState<boolean>(false);
  
  // サーバー側から ToDo リストを取得する
  const fetchData = async () => {
    try {
      const response = await fetch('http://localhost:1323/api/todos'); // APIエンドポイントを指定
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      const data = await response.json(); // JSON データをパース

      // content.id をインデックスとしてオブジェクトに変換
      const table = data.reduce((acc: ToDoTable, item: ToDoListItem) => {
        acc[item.content.id] = item;
        return acc;
      }, {});

      setTable(table);
      setIsInitialized(true);
    } catch (error) {
      console.error('Fetching error: ', error);
    }
  };

  // クライアント側で変更された table の状態をサーバーに送信する関数
  const sendData = async () => {
    try {
      const response = await fetch('http://localhost:1323/api/todos', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(Object.values(table)), // テーブルの状態をサーバーに送信
      });
      if (!response.ok) {
        throw new Error('Failed to send data to the server');
      }
    } catch (error) {
      console.error('Sending error: ', error);
    }
  };

  // コンポーネントマウント時にサーバーから ToDo リストを取得する
  useEffect(() => {
    if (!isInitialized) {
      fetchData(); // 最初の１回だけサーバーからデータをフェッチ
    } else {
      sendData(); // table が変更されたときにサーバーに送信
    }
  }, [table]);

  const updateTable = (id: number, title: string) => {
    const newTable = { ...table };
    newTable[id].content = { id, title };

    setTable(newTable);
    console.log(table); // デバッグ用
  };

  const addToDoCard = () => {
    const maxId = 
      (Object.keys(table).length == 0) ? 0 : Object.keys(table).map(Number).reduce((a, b) => Math.max(a, b));
      
    const newId = maxId + 1;
    const newKey = newId + 1;

    const newTable = { ...table };
    newTable[newId] = { key: newKey, content: { id: newId, title: 'New Task' } };

    setTable(newTable);
    console.log(table); // デバッグ用
  };

  const deleteToDoCard = (id: number) => {
    const newTable = { ...table };
    delete newTable[id];

    setTable(newTable);
    console.log(table); // デバッグ用
  };

  return (
    <>
      <div className='todo-grid-wrapper'>
        {Object.values(table).map(({key, content}) => (
          <ToDoCard
            key={key}
            id={content.id}
            title={content.title}
            updateTable={updateTable}
            deleteToDoCard={deleteToDoCard}
          />
        ))}
        <AddToDoCardButton addToDoCard={addToDoCard}/>
      </div>
    </>
  );
}

interface ToDoCardProps {
  id: number;
  title: string;
  updateTable: (id: number, title: string) => void;
  deleteToDoCard: (id: number) => void;
};

const ToDoCard: React.FC<ToDoCardProps> = ({ id, title, updateTable, deleteToDoCard }) => {
  // key は ToDoList が要素を管理するために React が使用する値なのでここでは受け取らない
  const [isEditing, setIsEditing] = useState(false);
  const [inputValue, setInputValue] = useState(title); // 入力値を状態管理

  const handleTitleClick = () => {
    setIsEditing(true); // クリックされたときに入力状態にする
  };
  
  const handleTitleBlur = () => {
    setIsEditing(false); // フォーカスが外れたときに再度 <div> に戻る

    if (inputValue == '') {
      // inputValue が空だと編集できなくなるので
      // 空の場合は埋め草を入れておく
      title = 'New Task';
      setInputValue(title);
      updateTable(id, title);
    } else {
      updateTable(id, inputValue);
    }
  };

  const handleDeleteButtonClick = () => {
    deleteToDoCard(id);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value); // 入力値をリアルタイムに状態に反映
  };

  const viewer = (
    <>
      <div className='todo-grid-item'>
        <button className='todo-grid-item-button' onClick={handleDeleteButtonClick}>-</button>
        <span className='todo-grid-item-title' onClick={handleTitleClick}>{inputValue}</span>
      </div>
    </>
  );

  const editor = (
    <>
      <div className='todo-grid-item'>
        <input
          className='todo-grid-item-input'
          type='text'
          value={inputValue}
          onChange={handleInputChange}
          autoFocus
          onBlur={handleTitleBlur}
        />
      </div>
    </>
  );

  return (
    <>
      { isEditing ? editor : viewer }
    </>
  );
}

interface AddToDoCardButtonProps {
  addToDoCard: () => void;
};

const AddToDoCardButton: React.FC<AddToDoCardButtonProps> = ({addToDoCard}) => {
  const handleOnClick = () => {
    addToDoCard();
  };

  return (
    <>
      <div className='todo-grid-item' onClick={handleOnClick}>
        <button className='todo-grid-item-button'>+</button>
      </div>
    </>
  )
}

function App() {
  return (
    <>
      <h1>Rabit ToDo</h1>
      <ToDoList/>
    </>
  );
}

export default App
