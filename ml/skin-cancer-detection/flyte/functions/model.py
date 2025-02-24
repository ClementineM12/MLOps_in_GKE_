# Import necessities
from glob import glob

import numpy as np
import pandas as pd
import PIL.Image as Image
import matplotlib.pyplot as plt
import seaborn as sns

from keras.layers import Dense, Flatten
from tensorflow.keras.applications import VGG16
from tensorflow.keras.layers import Dense, Flatten, Dropout
from tensorflow.keras.models import Model
from tensorflow.keras.optimizers import SGD
from tensorflow.keras.regularizers import l2

from sklearn.model_selection import train_test_split
from sklearn.decomposition import PCA
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score
from sklearn.metrics import confusion_matrix, classification_report
from collections import defaultdict

def train_random_forest(data, seed=42):
    skin_features = data[['perimeter', 'non_zeros', 'circularity',
       'main_assymetry', 'secondary_assymetry', 'r_channel', 'g_channel',
       'b_channel', 'label']]
    
    skin_features = skin_features.sample(frac=1, random_state=seed)

    skin_features.head()

    skin_features['perimeter'] = skin_features['perimeter'].astype(float)
    skin_features['non_zeros'] = skin_features['perimeter'].astype(float)
    skin_features['circularity'] = skin_features['perimeter'].astype(float)
    skin_features['main_assymetry'] = skin_features['perimeter'].astype(float)
    skin_features['secondary_assymetry'] = skin_features['perimeter'].astype(float)
    
    # create a PCA object with n_components=1
    pca = PCA(n_components=1)

    for channel in ['r_channel', 'g_channel', 'b_channel']:
        # Apply PCA to each row in the channel column and save the result in the same place
        skin_features[channel] = skin_features[channel].apply(lambda x: pca.fit_transform(x).flatten()[0])

    # Split the dataset into training and testing sets
    X_train, X_test, y_train, y_test = train_test_split(data, labels, test_size=0.25, random_state=seed)

    RF = RandomForest(RandomForestClassifier(random_state=seed), X_train, y_train, X_test, y_test, 'RF')


# Random Forest classifier
def RandomForest(model, X_train, y_train, X_test, y_test, title):
    predictions = defaultdict(list)

    model = model
    hist = model.fit(X_train, y_train)
    y_pred=model.predict(X_test)
    predictions[title].append(y_pred)
    print('Model accuracy score with 10 decision-trees: {0:0.4f}%'. format(accuracy_score(y_test, y_pred) * 100))
    
    target_names = ['Melanoma', 'Non-melanoma']
    print(classification_report(y_test, y_pred, target_names=target_names, zero_division=0))
    
    cm = confusion_matrix(y_test, y_pred)
    cm_matrix = pd.DataFrame(data=cm, columns=['Actual Normal', 'Actual Pathologic'], 
                                 index=['Predicted Normal', 'Predicted Pathologic'])
    sns.heatmap(cm_matrix, annot=True, fmt='d', cmap='coolwarm')
    print('Confusion matrix:\n')
    return model

# Function to plot loss during training
def plot_loss(model_fitting):
    """Plots the training and validation loss values of a deep learning model over epochs"""
    plt.plot(model_fitting.history['loss'], label='train') 
    if 'val_loss' in model_fitting.history:
        plt.plot(model_fitting.history['val_loss'], label='test')
    plt.title('Model Loss')
    plt.xlabel('epochs')
    plt.ylabel('loss values')
    plt.legend(loc='upper right')
    plt.show()
    
# Function to plot accuracy measurements during training-testing
def plot_accuracy(model_fitting):
    """Plots the training and validation accuracy values of a deep learning model over epochs"""
    plt.plot(model_fitting.history['accuracy'], label='train')
    if 'val_accuracy' in model_fitting.history:
        plt.plot(model_fitting.history['val_accuracy'], label='test')
    plt.title('Model Accuracy')
    plt.xlabel('epochs')
    plt.ylabel('accuracy')
    plt.legend(loc='upper right')
    plt.show()


# Define a function to load and preprocess images
def load_images(df, target_size=(224, 224)):
    images = []
    for i in range(len(df)):
        img_array = df.iloc[i]['image']
        img = Image.fromarray(img_array)
        img = img.resize(target_size)
        x = np.array(img)
        x = np.expand_dims(x, axis=0)
        images.append(x)
    return np.vstack(images)


def train_cnn(data, sample=400):
    skin_cnn = data[:sample]

    # Split the data into training, validation, and test sets
    df_train_val, df_test = train_test_split(skin_cnn, test_size=0.2, stratify=skin_cnn['label'])
    df_train, df_val = train_test_split(df_train_val, test_size=0.25, stratify=df_train_val['label'])

    # Load and preprocess the training, validation, and test images
    X_train = load_images(df_train)
    X_val = load_images(df_val)
    X_test = load_images(df_test)
    y_train = df_train['label']
    y_val = df_val['label']
    y_test = df_test['label']

    # Load a pre-trained VGG16 model and remove the top layers
    base_model = VGG16(weights='imagenet', include_top=False, input_shape=(224, 224, 3))
    x = base_model.output
    x = Flatten()(x)
    x = Dense(256, activation='relu', kernel_regularizer=l2(0.01))(x)
    x = Dropout(0.5)(x)
    x = Dense(128, activation='relu', kernel_regularizer=l2(0.01))(x)
    predictions = Dense(1, activation='sigmoid')(x)

    # Create the model
    model = Model(inputs=base_model.input, outputs=predictions)

    # Unfreeze the last 4 convolutional blocks and train with a lower learning rate
    for layer in base_model.layers[15:]:
        layer.trainable = True
    optimizer = SGD(lr=0.00001, momentum=0.9)
    model.compile(optimizer=optimizer, loss='binary_crossentropy', metrics=['accuracy'])

    # Train the model with validation data
    history = model.fit(X_train, y_train, batch_size=6, epochs=10, validation_data=(X_val, y_val))

    # Evaluate the model on the test set
    y_pred = model.predict(X_test)
    y_pred = (y_pred > 0.5).astype(int).ravel()
    accuracy = np.mean(y_pred == y_test)
    print('Test accuracy:', accuracy)

    plot_loss(history)
    plot_accuracy(history)

    return model